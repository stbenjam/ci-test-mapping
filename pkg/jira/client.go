package jira

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const SiteURL = "https://redhat.atlassian.net/"

// Client shares an HTTP client and base URL across a command's Jira reads.
type Client struct {
	httpClient    *http.Client
	baseURL       string
	authorization string
}

// NewClient uses the supplied HTTP client and base URL. authorization applies
// only to legacy clients; pass an empty string for an OAuth HTTP client.
func NewClient(httpClient *http.Client, baseURL, authorization string) *Client {
	return &Client{
		httpClient:    httpClient,
		baseURL:       strings.TrimRight(baseURL, "/") + "/",
		authorization: authorization,
	}
}

// NewRequest builds a GET with a relative REST path, preserving gateway prefixes.
func (c *Client) NewRequest(ctx context.Context, restPath string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+restPath, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if c.authorization != "" {
		req.Header.Set("Authorization", c.authorization)
	}
	return req, nil
}

// Do sends a Jira request through the shared HTTP client.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}

func (c *Client) read(req *http.Request) ([]byte, error) {
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}
