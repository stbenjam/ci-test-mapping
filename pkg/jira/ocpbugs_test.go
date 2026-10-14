package jira

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestGetJiraComponents(t *testing.T) {
	for _, prefix := range []string{"", "/ex/jira/test-cloud"} {
		t.Run(prefix, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != prefix+"/rest/api/2/project/OCPBUGS/components" || r.Header.Get("Authorization") != "Bearer legacy" || r.Method != http.MethodGet {
					t.Errorf("unexpected request: %s %s %v", r.Method, r.URL, r.Header)
				}
				if _, err := fmt.Fprint(w, `[{"id":"42","name":"Networking"},{"id":"invalid","name":"Skipped"},{"id":"43","name":"Documentation"}]`); err != nil {
					t.Error(err)
				}
			}))
			defer server.Close()
			client := NewClient(server.Client(), server.URL+prefix, "Bearer legacy")
			got, err := client.GetJiraComponents(context.Background())
			want := map[string]int64{"Networking": 42, "Documentation": 43}
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("components=%v, err=%v", got, err)
			}
		})
	}
}

func TestGetJiraComponentsMalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := fmt.Fprint(w, `{"unexpected":true}`); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()
	if _, err := NewClient(server.Client(), server.URL, "").GetJiraComponents(context.Background()); err == nil {
		t.Fatal("expected decode error")
	}
}

func TestComponentRetryCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		// Cancel after the first failed attempt, before the minute-long retry.
		cancel()
	}))
	defer server.Close()
	if _, err := NewClient(server.Client(), server.URL, "").GetJiraComponents(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
}

func TestRetryWaitCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- waitForRetry(ctx, time.Minute) }()
	// Give the retry wait a live context, then cancel it during the delay.
	timer := time.AfterFunc(10*time.Millisecond, cancel)
	defer timer.Stop()
	defer cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("retry wait ignored cancellation")
	}
}
