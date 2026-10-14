package flags

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync/atomic"
	"testing"

	"github.com/spf13/pflag"
	"golang.org/x/oauth2"
)

type jiraTestTransport struct {
	target *url.URL
	base   http.RoundTripper
}

func (r jiraTestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Scheme != "https" || (req.URL.Host != "auth.atlassian.com" && req.URL.Host != "api.atlassian.com" && req.URL.Host != "redhat.atlassian.net") {
		return nil, fmt.Errorf("unexpected Jira destination: %s", req.URL)
	}
	clone := req.Clone(req.Context())
	clone.URL.Scheme, clone.URL.Host = r.target.Scheme, r.target.Host
	clone.Host = r.target.Host
	return r.base.RoundTrip(clone)
}

func clearJiraEnv(t *testing.T) {
	t.Helper()
	for _, name := range []string{"JIRA_TOKEN", "JIRA_TOKEN_BASIC", "JIRA_CLIENT_ID", "JIRA_CLIENT_SECRET", "JIRA_CLOUD_ID"} {
		t.Setenv(name, "")
		if err := os.Unsetenv(name); err != nil {
			t.Fatal(err)
		}
	}
}

func TestJiraAuthenticationSelection(t *testing.T) {
	const cloudID = "35273b54-3f06-40d2-880f-dd28cf8daafa"
	for _, tc := range []struct {
		name       string
		oauth      bool
		required   bool
		env        map[string]string
		wantAuth   string
		wantID     string
		wantSecret string
		wantError  bool
		reject     bool
	}{
		{name: "anonymous map"},
		{name: "missing required PAT", required: true, wantError: true},
		{name: "bearer precedence", required: true, env: map[string]string{"JIRA_TOKEN": "pat", "JIRA_TOKEN_BASIC": "basic"}, wantAuth: "Bearer pat"},
		{name: "basic", required: true, env: map[string]string{"JIRA_TOKEN_BASIC": "basic"}, wantAuth: "Basic basic"},
		{name: "empty bearer uses basic", env: map[string]string{"JIRA_TOKEN": "", "JIRA_TOKEN_BASIC": "basic"}, wantAuth: "Basic basic"},
		{name: "OAuth variables do not opt in", env: map[string]string{"JIRA_CLIENT_ID": "client", "JIRA_CLIENT_SECRET": "secret", "JIRA_CLOUD_ID": cloudID}},
		{name: "OAuth variables cannot replace required PAT", required: true, env: map[string]string{"JIRA_CLIENT_ID": "client", "JIRA_CLIENT_SECRET": "secret", "JIRA_CLOUD_ID": cloudID}, wantError: true},
		{name: "invalid OAuth ignored without flag", env: map[string]string{"JIRA_CLIENT_ID": "", "JIRA_CLOUD_ID": "invalid", "JIRA_TOKEN": "pat"}, wantAuth: "Bearer pat"},
		{name: "OAuth pair beats PAT", oauth: true, required: true, env: map[string]string{"JIRA_CLIENT_ID": "client", "JIRA_CLIENT_SECRET": "secret", "JIRA_CLOUD_ID": cloudID, "JIRA_TOKEN": "pat", "JIRA_TOKEN_BASIC": "basic"}, wantID: "client", wantSecret: "secret"},
		{name: "missing client ID", oauth: true, env: map[string]string{"JIRA_CLIENT_SECRET": "secret", "JIRA_CLOUD_ID": cloudID}, wantError: true},
		{name: "missing client secret", oauth: true, env: map[string]string{"JIRA_CLIENT_ID": "client", "JIRA_CLOUD_ID": cloudID}, wantError: true},
		{name: "empty client ID", oauth: true, env: map[string]string{"JIRA_CLIENT_ID": "", "JIRA_CLIENT_SECRET": "secret", "JIRA_CLOUD_ID": cloudID}, wantError: true},
		{name: "empty client secret", oauth: true, env: map[string]string{"JIRA_CLIENT_ID": "client", "JIRA_CLIENT_SECRET": "", "JIRA_CLOUD_ID": cloudID}, wantError: true},
		{name: "empty credential pair", oauth: true, env: map[string]string{"JIRA_CLIENT_ID": "", "JIRA_CLIENT_SECRET": "", "JIRA_CLOUD_ID": cloudID}, wantError: true},
		{name: "invalid cloud fails despite PAT", oauth: true, env: map[string]string{"JIRA_CLIENT_ID": "client", "JIRA_CLIENT_SECRET": "secret", "JIRA_CLOUD_ID": "not-a-uuid", "JIRA_TOKEN": "pat"}, wantError: true},
		{name: "missing cloud", oauth: true, env: map[string]string{"JIRA_CLIENT_ID": "client", "JIRA_CLIENT_SECRET": "secret"}, wantError: true},
		{name: "missing OAuth credentials fails despite PAT", oauth: true, env: map[string]string{"JIRA_TOKEN": "pat"}, wantError: true},
		{name: "rejected grant never falls back", oauth: true, env: map[string]string{"JIRA_CLIENT_ID": "client", "JIRA_CLIENT_SECRET": "secret", "JIRA_CLOUD_ID": cloudID, "JIRA_TOKEN_BASIC": "basic"}, wantID: "client", wantSecret: "secret", reject: true, wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clearJiraEnv(t)
			for key, value := range tc.env {
				t.Setenv(key, value)
			}
			var grants, reads atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/oauth/token" {
					grants.Add(1)
					if err := r.ParseForm(); err != nil {
						t.Error(err)
					}
					if r.PostForm.Get("client_id") != tc.wantID || r.PostForm.Get("client_secret") != tc.wantSecret {
						t.Errorf("wrong credential pair: %v", r.PostForm)
					}
					w.Header().Set("Content-Type", "application/json")
					if tc.reject {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					if _, err := fmt.Fprint(w, `{"access_token":"oauth","token_type":"Bearer","expires_in":3600}`); err != nil {
						t.Error(err)
					}
					return
				}
				reads.Add(1)
				wantPath, wantAuth := "/rest/api/2/project/OCPBUGS/components", tc.wantAuth
				if tc.oauth {
					wantPath, wantAuth = "/ex/jira/"+cloudID+wantPath, "Bearer oauth"
				}
				if r.URL.Path != wantPath || r.Header.Get("Authorization") != wantAuth {
					t.Errorf("request = %s, %q; want %s, %q", r.URL.Path, r.Header.Get("Authorization"), wantPath, wantAuth)
				}
			}))
			defer server.Close()
			target, err := url.Parse(server.URL)
			if err != nil {
				t.Fatal(err)
			}
			transport := jiraTestTransport{target: target, base: server.Client().Transport}
			originalTransport := http.DefaultTransport
			http.DefaultTransport = transport
			t.Cleanup(func() { http.DefaultTransport = originalTransport })
			ctx := context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{Transport: transport})
			client, err := (&JiraFlags{ServiceAccount: tc.oauth}).GetClient(ctx, tc.required)
			if tc.wantError {
				if err == nil || client != nil || reads.Load() != 0 {
					t.Fatalf("GetClient = %v, %v; reads=%d", client, err, reads.Load())
				}
				if !tc.reject && grants.Load() != 0 {
					t.Fatal("invalid configuration attempted token grant")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			for range 2 {
				req, err := client.NewRequest(context.Background(), "rest/api/2/project/OCPBUGS/components")
				if err != nil {
					t.Fatal(err)
				}
				wantHost := "redhat.atlassian.net"
				if tc.oauth {
					wantHost = "api.atlassian.com"
					if req.Header.Get("Authorization") != "" {
						t.Fatal("legacy header attached to OAuth request")
					}
				}
				if req.URL.Host != wantHost {
					t.Fatalf("host=%s, want %s", req.URL.Host, wantHost)
				}
				resp, err := client.Do(req)
				if err != nil {
					t.Fatal(err)
				}
				if err := resp.Body.Close(); err != nil {
					t.Fatal(err)
				}
			}
			wantGrants := int32(0)
			if tc.oauth {
				wantGrants = 1
			}
			if grants.Load() != wantGrants || reads.Load() != 2 {
				t.Fatalf("grants=%d reads=%d", grants.Load(), reads.Load())
			}
		})
	}
}

func TestJiraFlagDefault(t *testing.T) {
	var f JiraFlags
	fs := pflag.NewFlagSet("jira", pflag.ContinueOnError)
	f.BindFlags(fs)
	if f.ServiceAccount || fs.Lookup("jira-service-account").DefValue != "false" {
		t.Fatal("service-account auth must default to false")
	}
	if err := fs.Parse([]string{"--jira-service-account"}); err != nil {
		t.Fatal(err)
	}
	if !f.ServiceAccount {
		t.Fatal("flag did not opt in")
	}
}
