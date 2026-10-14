package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/oauth2"

	"github.com/openshift-eng/ci-test-mapping/pkg/jira"
)

func TestCreationMetadataResponses(t *testing.T) {
	for _, tc := range []struct {
		name   string
		types  string
		fields string
		want   []string
		err    string
	}{
		{"site format", `{"issueTypes":[{"id":"9","name":"Task"},{"id":"17","name":"Bug"}]}`, `{"fields":[{"fieldId":"other","allowedValues":[{"name":"Ignored"}]},{"fieldId":"components","allowedValues":[{"name":"Networking"},{"name":"Documentation"},{"name":"API Documentation"},{"name":"Storage"}]}]}`, []string{"Networking", "Storage"}, ""},
		{"values format", `{"values":[{"id":"17","name":"Bug"}]}`, `{"values":[{"fieldId":"components","allowedValues":[{"name":"Networking"}]}]}`, []string{"Networking"}, ""},
		{"prefer issueTypes and fields", `{"issueTypes":[{"id":"17","name":"Bug"}],"values":[{"id":"99","name":"Bug"}]}`, `{"fields":[{"fieldId":"components","allowedValues":[{"name":"Networking"}]}],"values":[{"fieldId":"components","allowedValues":[{"name":"Ignored"}]}]}`, []string{"Networking"}, ""},
		{"no components", `{"values":[{"id":"17","name":"Bug"}]}`, `{"values":[]}`, nil, ""},
		{"missing bug", `{"issueTypes":[{"id":"9","name":"Task"}]}`, "", nil, "jira bug type id required"},
		{"invalid issue types", `{"values":`, "", nil, "could not fetch jira bug type"},
		{"invalid fields", `{"issueTypes":[{"id":"17","name":"Bug"}]}`, `{"fields":`, nil, "unexpected end"},
	} {
		for _, prefix := range []string{"", "/ex/jira/test-cloud"} {
			t.Run(tc.name+prefix, func(t *testing.T) {
				var paths []string
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					paths = append(paths, r.URL.Path)
					if r.Method != http.MethodGet || r.Header.Get("Authorization") != "Basic legacy" || r.Header.Get("Accept") != "application/json" {
						t.Errorf("unexpected metadata request: %s %v", r.Method, r.Header)
					}
					body := tc.types
					if strings.HasSuffix(r.URL.Path, "/17") {
						body = tc.fields
					} else if r.URL.Path != prefix+"/"+createMetadataPath {
						t.Errorf("unexpected metadata path %q", r.URL.Path)
					}
					if _, err := fmt.Fprint(w, body); err != nil {
						t.Error(err)
					}
				}))
				defer server.Close()
				client := jira.NewClient(server.Client(), server.URL+prefix, "Basic legacy")
				got, err := getJiraComponents(context.Background(), client)
				if tc.err != "" {
					if err == nil || !strings.Contains(err.Error(), tc.err) {
						t.Fatalf("expected %q, got %v", tc.err, err)
					}
					return
				}
				if err != nil || !reflect.DeepEqual(got, tc.want) {
					t.Fatalf("components=%v, err=%v", got, err)
				}
				if !reflect.DeepEqual(paths, []string{prefix + "/" + createMetadataPath, prefix + "/" + createMetadataPath + "17"}) {
					t.Fatalf("paths=%v", paths)
				}
			})
		}
	}
}

type commandJiraTransport struct {
	target *url.URL
	base   http.RoundTripper
}

func (r commandJiraTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Scheme != "https" || (req.URL.Host != "auth.atlassian.com" && req.URL.Host != "api.atlassian.com") {
		return nil, fmt.Errorf("unexpected command Jira destination: %s", req.URL)
	}
	clone := req.Clone(req.Context())
	clone.URL.Scheme, clone.URL.Host = r.target.Scheme, r.target.Host
	clone.Host = r.target.Host
	return r.base.RoundTrip(clone)
}

func setupLocalMapping(t *testing.T) {
	t.Helper()
	t.Chdir(t.TempDir())
	for filename, body := range map[string]string{
		"data/openshift-gce-devel/ci_analysis_us/junit.json": "[]",
		"pkg/components/example/component.go":                "package example\nvar ExampleComponent = \"Example\"\n",
		"pkg/registry/registry.go":                           "package registry\nimport (\n)\nfunc register() {\n// New components go here\n}\n",
	} {
		if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filename, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	saved := f
	f = NewMapFlags()
	f.BindFlags(pflag.NewFlagSet("map", pflag.ContinueOnError))
	f.mode, f.junitTable, f.testMappingTable = ModeLocal, "junit", "component_mapping"
	f.bigqueryFlags.ServiceAccountCredentialFile = ""
	// The existing map command's flags bind the original f, so configure the
	// temporary flags directly through their own FlagSet.
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	f.configFlags.BindFlags(fs)
	if err := fs.Set("config", ""); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f = saved })
}

func TestJiraCommandsWithOAuth(t *testing.T) {
	const cloudID = "35273b54-3f06-40d2-880f-dd28cf8daafa"
	for _, command := range []*cobra.Command{mapCmd, createCmd, verifyCmd} {
		t.Run(command.Name(), func(t *testing.T) {
			setupLocalMapping(t)
			t.Setenv("JIRA_CLIENT_ID", "client")
			t.Setenv("JIRA_CLIENT_SECRET", "secret")
			t.Setenv("JIRA_CLOUD_ID", cloudID)
			t.Setenv("JIRA_TOKEN", "legacy-pat")
			t.Setenv("JIRA_TOKEN_BASIC", "legacy-basic")
			// Each command exposes the same opt-in flag, defaulting to false.
			flag := command.Flags().Lookup("jira-service-account")
			if flag == nil || flag.DefValue != "false" {
				t.Fatal("missing default-false Jira flag")
			}
			if command == mapCmd {
				f.jiraFlags.ServiceAccount = true
			} else {
				if err := command.Flags().Set("jira-service-account", "true"); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if err := command.Flags().Set("jira-service-account", "false"); err != nil {
						t.Error(err)
					}
				})
			}
			var grants, reads atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				body := ""
				if r.URL.Path == "/oauth/token" {
					grants.Add(1)
					body = `{"access_token":"command-token","token_type":"Bearer","expires_in":3600}`
				} else {
					reads.Add(1)
					if r.Method != http.MethodGet || r.Header.Get("Authorization") != "Bearer command-token" || len(r.Header.Values("Authorization")) != 1 {
						t.Errorf("wrong command auth or method: %s, %v", r.Method, r.Header)
					}
					switch r.URL.Path {
					case "/ex/jira/" + cloudID + "/rest/api/2/project/OCPBUGS/components":
						body = `[{"id":"42","name":"Networking"}]`
					case "/ex/jira/" + cloudID + "/" + createMetadataPath:
						body = `{"issueTypes":[{"id":"17","name":"Bug"}]}`
					case "/ex/jira/" + cloudID + "/" + createMetadataPath + "17":
						body = `{"fields":[{"fieldId":"components","allowedValues":[{"name":"OAuth Test Fixture"}]}]}`
					default:
						t.Errorf("unexpected command path: %s", r.URL.Path)
					}
				}
				if _, err := fmt.Fprint(w, body); err != nil {
					t.Error(err)
				}
			}))
			defer server.Close()
			target, err := url.Parse(server.URL)
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{
				Transport: commandJiraTransport{target: target, base: server.Client().Transport},
			})
			command.SetContext(ctx)
			t.Cleanup(func() { command.SetContext(context.Background()) })
			if err := command.RunE(command, nil); err != nil {
				t.Fatal(err)
			}
			wantReads := int32(2)
			if command == mapCmd {
				wantReads = 1
				data, err := os.ReadFile("data/openshift-gce-devel/ci_analysis_us/component_mapping.json")
				if err != nil || strings.TrimSpace(string(data)) != "null" {
					t.Fatalf("map local output: %q, %v", data, err)
				}
			}
			if grants.Load() != 1 || reads.Load() != wantReads {
				t.Fatalf("grants=%d reads=%d", grants.Load(), reads.Load())
			}
			if command == createCmd {
				data, err := os.ReadFile("pkg/components/oauthtestfixture/component.go")
				if err != nil || !strings.Contains(string(data), `var OAuthTestFixtureComponent = "OAuth Test Fixture"`) {
					t.Fatalf("jira-create local output: %q, %v", data, err)
				}
				data, err = os.ReadFile("pkg/registry/registry.go")
				if err != nil || !strings.Contains(string(data), `r.Register("OAuth Test Fixture", &oauthtestfixture.OAuthTestFixtureComponent)`) {
					t.Fatalf("jira-create registry: %q, %v", data, err)
				}
			}
			canceled, cancel := context.WithCancel(ctx)
			cancel()
			client := jira.NewClient(server.Client(), server.URL, "")
			if _, err := getJiraComponents(canceled, client); !errors.Is(err, context.Canceled) {
				t.Fatalf("metadata request ignored cancellation: %v", err)
			}
		})
	}
}
