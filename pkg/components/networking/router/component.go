package networkingrouter

import (
	v1 "github.com/openshift-eng/ci-test-mapping/pkg/api/types/v1"
	"github.com/openshift-eng/ci-test-mapping/pkg/config"
	"github.com/openshift-eng/ci-test-mapping/pkg/util"
)

type Component struct {
	*config.Component
}

var RouterComponent = Component{
	Component: &config.Component{
		Name:                 "Networking / router",
		Operators:            []string{"ingress"},
		DefaultJiraComponent: "Networking / router",
		Matchers: []config.ComponentMatcher{
			{
				IncludeAll: []string{"bz-Routing"},
			},
			{
				SIG:        "sig-network",
				IncludeAll: []string{"Feature:Router"},
			},
			{
				SIG:        "sig-network-edge",
				IncludeAll: []string{"Feature:Router"},
			},
			{
				IncludeAll: []string{"ingress-to-", "disruption"},
			},
			{
				IncludeAll: []string{"openshift-ingress"},
			},
			{
				IncludeAll: []string{"via cluster ingress"},
			},
			{
				IncludeAll: []string{"Cluster frontend ingress"},
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
	return util.StableID(test)
}

func (c *Component) JiraComponents() (components []string) {
	components = []string{c.DefaultJiraComponent}
	for _, m := range c.Matchers {
		components = append(components, m.JiraComponent)
	}

	return components
}
