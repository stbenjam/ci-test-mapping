package flags

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/pflag"

	"github.com/openshift-eng/ci-test-mapping/pkg/jira"
	"github.com/openshift-eng/ci-test-mapping/pkg/jiraauth"
)

// JiraFlags selects authentication independently of mapping output options.
type JiraFlags struct {
	ServiceAccount bool
}

func (f *JiraFlags) BindFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&f.ServiceAccount, "jira-service-account", false, "Use Jira service-account OAuth credentials")
}

// GetClient constructs one client per command. ctx must remain alive for the
// command's lifetime, including retries and token renewals. Legacy commands may
// require credentials; map continues to permit unauthenticated requests.
func (f *JiraFlags) GetClient(ctx context.Context, requireCredentials bool) (*jira.Client, error) {
	if f.ServiceAccount {
		config := jiraauth.Config{
			ClientID:     os.Getenv("JIRA_CLIENT_ID"),
			ClientSecret: os.Getenv("JIRA_CLIENT_SECRET"),
			CloudID:      os.Getenv("JIRA_CLOUD_ID"),
		}
		httpClient, baseURL, err := config.NewClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("configure Jira service-account client (JIRA_CLIENT_ID, JIRA_CLIENT_SECRET, JIRA_CLOUD_ID): %w", err)
		}
		return jira.NewClient(httpClient, baseURL, ""), nil
	}

	var authorization string
	if token := os.Getenv("JIRA_TOKEN"); token != "" {
		authorization = "Bearer " + token
	} else if token := os.Getenv("JIRA_TOKEN_BASIC"); token != "" {
		authorization = "Basic " + token
	}
	if requireCredentials && authorization == "" {
		return nil, fmt.Errorf("jira token required")
	}
	return jira.NewClient(&http.Client{}, jira.SiteURL, authorization), nil
}
