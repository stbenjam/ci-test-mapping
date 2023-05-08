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
				IncludeSubstrings: []string{"bz-Cloud Compute"},
			},
			{
				IncludeSubstrings: []string{"bz-cluster-api"},
			},
			{
				IncludeSubstrings: []string{"bz-control-plane-machine-set"},
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
