// Package jiraauth provides service-account authentication for Jira HTTP callers.
package jiraauth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// Config identifies a service-account credential and the Jira cloud it accesses.
// Scopes are assigned to the credential in Atlassian Administration.
type Config struct {
	ClientID     string
	ClientSecret string `json:"-"`
	CloudID      string
}

// NewClient obtains an initial token and returns an authenticated client and
// gateway base URL. Keep ctx alive for the command or job's lifetime: the OAuth
// library retains it for replacement token grants. API requests use their own
// contexts. The token source caches tokens and replaces them when they expire.
func (c Config) NewClient(ctx context.Context) (*http.Client, string, error) {
	if c.ClientID == "" {
		return nil, "", fmt.Errorf("jira service-account client ID is required")
	}
	if c.ClientSecret == "" {
		return nil, "", fmt.Errorf("jira service-account client secret is required")
	}
	cloudID, err := uuid.Parse(c.CloudID)
	if err != nil {
		return nil, "", fmt.Errorf("jira service-account cloud ID must be a UUID: %w", err)
	}

	config := clientcredentials.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		TokenURL:     "https://auth.atlassian.com/oauth/token",
		AuthStyle:    oauth2.AuthStyleInParams,
	}
	source := config.TokenSource(ctx)
	if _, err := source.Token(); err != nil {
		return nil, "", fmt.Errorf("obtain Jira service-account access token: %w", err)
	}
	return oauth2.NewClient(ctx, source), "https://api.atlassian.com/ex/jira/" + cloudID.String() + "/", nil
}
