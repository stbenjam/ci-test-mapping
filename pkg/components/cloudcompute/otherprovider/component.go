package cloudcomputeotherprovider

import (
	v1 "github.com/openshift-eng/ci-test-mapping/pkg/api/types/v1"
	"github.com/openshift-eng/ci-test-mapping/pkg/config"
	"github.com/openshift-eng/ci-test-mapping/pkg/util"
)

type Component struct {
	*config.Component
}

var OtherProviderComponent = Component{
	Component: &config.Component{
		Name:                 "Cloud Compute / Other Provider",
		Operators:            []string{"cluster-api", "machine-approver", "machine-api", "control-plane-machine-set"},
		DefaultJiraComponent: "Cloud Compute / Other Provider",
		Matchers: []config.ComponentMatcher{
			{
				Include: []string{"bz-Cloud Compute"},
			},
			{
				Include: []string{"bz-cluster-api"},
			},
			{
				Include: []string{"bz-control-plane-machine-set"},
			},
			{
				Include: []string{"service-load-balancer-", "disruption"},
			},
			{
				SIG:     "sig-network-edge",
				Include: []string{"Application behind service load balancer"},
			},
		},
	},
}

func (c *Component) IdentifyTest(test *v1.TestInfo) (*v1.TestOwnership, error) {
	if matcher := c.FindMatch(test); matcher != nil {
		jira := matcher.JiraComponent
		if jira == "" {
			jira = c.DefaultJiraComponent
		}
		return &v1.TestOwnership{
			Name:          test.Name,
			Component:     c.Name,
			JIRAComponent: jira,
			Priority:      matcher.Priority,
			Capabilities:  append(matcher.Capabilities, identifyCapabilities(test)...),
		}, nil
	}

	return nil, nil
}

func (c *Component) StableID(test *v1.TestInfo) string {
	return util.StableID(test, nil)
}

func (c *Component) JiraComponents() (components []string) {
	components = []string{c.DefaultJiraComponent}
	for _, m := range c.Matchers {
		components = append(components, m.JiraComponent)
	}

	return components
}
