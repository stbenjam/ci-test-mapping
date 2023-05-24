package clusterversionoperator

import (
	v1 "github.com/openshift-eng/ci-test-mapping/pkg/api/types/v1"
	"github.com/openshift-eng/ci-test-mapping/pkg/config"
)

type Component struct {
	*config.Component
}

var ClusterVersionOperatorComponent = Component{
	Component: &config.Component{
		Name:                 "Cluster Version Operator",
		Operators:            []string{"version"},
		DefaultJiraComponent: "Cluster Version Operator",
		Matchers: []config.ComponentMatcher{
			{
				Include: []string{"cluster-version-operator"},
			},
			{
				Include: []string{"bz-Cluster Version Operator"},
			},
			{
				Include: []string{"upgrade"},
				// Let others claim upgrade tests (i.e. for their component)
				Priority: -10,
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
	return test.Name
}

func (c *Component) JiraComponents() (components []string) {
	components = []string{c.DefaultJiraComponent}
	for _, m := range c.Matchers {
		components = append(components, m.JiraComponent)
	}

	return components
}
