package jira

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	v1 "github.com/openshift-eng/ci-test-mapping/pkg/api/types/v1"
)

func GetJiraComponents(authorization string) (map[string]int64, error) {
	return NewClient(&http.Client{}, SiteURL, authorization).GetJiraComponents(context.Background())
}

// GetJiraComponents loads OCPBUGS component IDs using the command's shared client.
func (c *Client) GetJiraComponents(ctx context.Context) (map[string]int64, error) {
	start := time.Now()
	log.Infof("loading jira ocpbugs component information...")
	body, err := c.jiraRequest(ctx, "rest/api/2/project/OCPBUGS/components")
	if err != nil {
		return nil, err
	}

	var components []v1.JiraComponent
	err = json.Unmarshal(body, &components)
	if err != nil {
		return nil, errors.WithMessage(err, "response from jira: "+string(body))
	}

	ids := make(map[string]int64)
	for _, c := range components {
		jiraID, err := strconv.ParseInt(c.ID, 10, 64)
		if err != nil {
			log.WithError(err).Warnf("error parsing jira ID '%s'", c.ID)
			continue
		}

		ids[c.Name] = jiraID
	}

	log.Infof("jira ocpbugs components loaded in %+v", time.Since(start))
	return ids, nil
}

func (c *Client) jiraRequest(ctx context.Context, restPath string) ([]byte, error) {
	req, err := c.NewRequest(ctx, restPath)
	if err != nil {
		return nil, err
	}

	var finalError error // keep the last error seen in case all attempts fail
	for _, wait := range []time.Duration{0, 1, 5, 15} {
		if err := waitForRetry(ctx, wait*time.Minute); err != nil {
			return nil, err
		}

		var bytes []byte
		bytes, finalError = c.read(req)
		if finalError == nil {
			return bytes, nil
		}
		log.Errorf("jira request failed: %v", finalError)
	}
	return nil, finalError
}

func waitForRetry(ctx context.Context, delay time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if delay == 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
