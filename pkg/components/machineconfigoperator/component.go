package machineconfigoperator

import (
	v1 "github.com/openshift-eng/ci-test-mapping/pkg/api/types/v1"
	"github.com/openshift-eng/ci-test-mapping/pkg/config"
	"github.com/openshift-eng/ci-test-mapping/pkg/util"
)

type Component struct {
	*config.Component
}

var MachineConfigOperatorComponent = Component{
	Component: &config.Component{
		Name:                 "Machine Config Operator",
		Operators:            []string{"machine-config"},
		DefaultJiraComponent: "Machine Config Operator",
		Matchers: []config.ComponentMatcher{
			{
				Include: []string{"bz-Machine Config Operator"},
			},
			{
				Include: []string{"bz-machine config operator"},
			},
			{
				Include: []string{"machine-config-operator"},
			},
			{
				SIG:          "sig-cluster-lifecycle",
				Include:      []string{"Pods cannot access the /config"},
				Capabilities: []string{"Config"},
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
