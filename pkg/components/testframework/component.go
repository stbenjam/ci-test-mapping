package testframework

import (
	v1 "github.com/openshift-eng/ci-test-mapping/pkg/api/types/v1"
	"github.com/openshift-eng/ci-test-mapping/pkg/config"
	"github.com/openshift-eng/ci-test-mapping/pkg/util"
)

type Component struct {
	*config.Component
}

var TestFrameworkComponent = Component{
	Component: &config.Component{
		Name:                 "Test Framework",
		Operators:            []string{},
		DefaultJiraComponent: "Test Framework",
		Matchers: []config.ComponentMatcher{
			{
				IncludeSubstrings: []string{"Undiagnosed panic"},
			},
			{
				SIG: "sig-trt",
			},
			{
				SIG: "sig-ci",
			},
			{
				IncludeSubstrings: []string{"ci-cluster-network-liveness-", "-connections"},
				Capabilities:      []string{"Build Clusters"},
			},

			// TRT is owner of last resort for these
			{
				SIG:               "sig-arch",
				IncludeSubstrings: []string{"events should not repeat"},
				Capabilities:      []string{"Pathological Events"},
				Priority:          -1,
			},
			{
				SIG:               "sig-arch",
				IncludeSubstrings: []string{"Alerts alert/"},
				Capabilities:      []string{"Alerts"},
				Priority:          -1,
			},

			{
				IncludeSubstrings: []string{"bz-Unknown", "alert/KubePodNotReady"},
				Priority:          -1,
				Capabilities:      []string{"KubePodNotReady - Other"},
			},
		},
	},
}

func (c *Component) IdentifyTest(test *v1.TestInfo) (*v1.TestOwnership, error) {
	if matcher := c.FindMatch(test); matcher != nil {
		return &v1.TestOwnership{
			Name:          test.Name,
			Component:     c.Name,
			JIRAComponent: matcher.JiraComponent,
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
