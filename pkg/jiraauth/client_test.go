package jiraauth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"golang.org/x/oauth2"
)

const testCloudID = "35273b54-3f06-40d2-880f-dd28cf8daafa"

type redirectedTransport struct {
	target *url.URL
	base   http.RoundTripper
}

func (r redirectedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Scheme != "https" || (req.URL.Host != "auth.atlassian.com" && req.URL.Host != "api.atlassian.com") {
		return nil, fmt.Errorf("unexpected destination: %s", req.URL)
	}
	clone := req.Clone(req.Context())
	clone.URL.Scheme, clone.URL.Host = r.target.Scheme, r.target.Host
	clone.Host = r.target.Host
	return r.base.RoundTrip(clone)
}

func localContext(t *testing.T, server *httptest.Server) context.Context {
	t.Helper()
	target, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	return context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{
		Transport: redirectedTransport{target: target, base: server.Client().Transport},
	})
}

func TestNewClientValidation(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config Config
		want   string
	}{
		{"missing ID", Config{ClientSecret: "secret", CloudID: testCloudID}, "client ID"},
		{"missing secret", Config{ClientID: "client", CloudID: testCloudID}, "client secret"},
		{"missing cloud", Config{ClientID: "client", ClientSecret: "secret"}, "UUID"},
		{"invalid cloud", Config{ClientID: "client", ClientSecret: "secret", CloudID: "../../other"}, "UUID"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client, baseURL, err := tc.config.NewClient(context.Background())
			if err == nil || !strings.Contains(err.Error(), tc.want) || client != nil || baseURL != "" {
				t.Fatalf("NewClient = %v, %q, %v", client, baseURL, err)
			}
		})
	}
}

func TestNewClientTokenReuseAndRenewal(t *testing.T) {
	for _, expires := range []int{3600, 1} {
		t.Run(fmt.Sprintf("expires_in=%d", expires), func(t *testing.T) {
			var grants, reads atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Path == "/oauth/token" {
					if err := r.ParseForm(); err != nil {
						t.Error(err)
					}
					if r.Method != http.MethodPost || r.Header.Get("Authorization") != "" ||
						r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" ||
						r.PostForm.Get("grant_type") != "client_credentials" ||
						r.PostForm.Get("client_id") != "client" || r.PostForm.Get("client_secret") != "secret&+" || len(r.PostForm) != 3 {
						t.Errorf("invalid token request: %s, %v, %v", r.Method, r.Header, r.PostForm)
					}
					grant := grants.Add(1)
					ttl := 3600
					if grant == 1 {
						ttl = expires
					}
					if _, err := fmt.Fprintf(w, `{"access_token":"token-%d","token_type":"Bearer","expires_in":%d}`, grant, ttl); err != nil {
						t.Error(err)
					}
					return
				}
				reads.Add(1)
				if r.URL.Path != "/ex/jira/"+testCloudID+"/rest/api/2/project/OCPBUGS/components" ||
					r.Method != http.MethodGet || r.Header.Get("Authorization") != fmt.Sprintf("Bearer token-%d", grants.Load()) || len(r.Header.Values("Authorization")) != 1 {
					t.Errorf("invalid Jira request: %s, %s, %v", r.Method, r.URL, r.Header)
				}
			}))
			defer server.Close()
			client, baseURL, err := (Config{ClientID: "client", ClientSecret: "secret&+", CloudID: testCloudID}).NewClient(localContext(t, server))
			if err != nil {
				t.Fatal(err)
			}
			if grants.Load() != 1 || reads.Load() != 0 || baseURL != "https://api.atlassian.com/ex/jira/"+testCloudID+"/" {
				t.Fatal("initial grant or gateway URL incorrect")
			}
			for range 3 {
				resp, err := client.Get(baseURL + "rest/api/2/project/OCPBUGS/components")
				if err != nil {
					t.Fatal(err)
				}
				if err := resp.Body.Close(); err != nil {
					t.Fatal(err)
				}
			}
			wantGrants := int32(1)
			if expires == 1 {
				// oauth2 considers tokens expired ten seconds early; no sleep needed.
				wantGrants = 2
			}
			if grants.Load() != wantGrants || reads.Load() != 3 {
				t.Fatalf("grants=%d, reads=%d", grants.Load(), reads.Load())
			}
		})
	}
}

func TestNewClientInitialGrantErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
	}{
		{"rejected credential", http.StatusUnauthorized, `{"error":"invalid_client"}`},
		{"missing access token", http.StatusOK, `{"token_type":"Bearer"}`},
		{"invalid JSON", http.StatusOK, `{"access_token":`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if _, err := fmt.Fprint(w, tc.body); err != nil {
					t.Error(err)
				}
			}))
			defer server.Close()
			client, baseURL, err := (Config{ClientID: "client", ClientSecret: "secret", CloudID: testCloudID}).NewClient(localContext(t, server))
			if err == nil || client != nil || baseURL != "" {
				t.Fatalf("NewClient = %v, %q, %v", client, baseURL, err)
			}
		})
	}
}

func TestClientLifetimeCancellation(t *testing.T) {
	for _, expires := range []int{3600, 1} {
		t.Run(fmt.Sprintf("expires_in=%d", expires), func(t *testing.T) {
			var grants, reads atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/oauth/token" {
					grants.Add(1)
					w.Header().Set("Content-Type", "application/json")
					if _, err := fmt.Fprintf(w, `{"access_token":"initial","token_type":"Bearer","expires_in":%d}`, expires); err != nil {
						t.Error(err)
					}
					return
				}
				reads.Add(1)
			}))
			defer server.Close()
			ctx, cancel := context.WithCancel(localContext(t, server))
			defer cancel()
			client, baseURL, err := (Config{ClientID: "client", ClientSecret: "secret", CloudID: testCloudID}).NewClient(ctx)
			if err != nil {
				t.Fatal(err)
			}
			cancel()
			// API context is independent; renewal must use the canceled lifetime.
			resp, err := client.Get(baseURL + "rest/api/2/project/OCPBUGS/components")
			if expires == 1 {
				if !errors.Is(err, context.Canceled) || resp != nil || reads.Load() != 0 {
					t.Fatalf("canceled renewal: %v, %v, reads=%d", resp, err, reads.Load())
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if err := resp.Body.Close(); err != nil {
					t.Fatal(err)
				}
				if reads.Load() != 1 {
					t.Fatal("cached valid token was not usable with independent API context")
				}
			}
			if grants.Load() != 1 {
				t.Fatal("token grant occurred after cancellation")
			}
			if _, _, err := (Config{ClientID: "client", ClientSecret: "secret", CloudID: testCloudID}).NewClient(ctx); !errors.Is(err, context.Canceled) {
				t.Fatalf("initial canceled grant: %v", err)
			}
		})
	}
}
