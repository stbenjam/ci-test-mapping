package monitoring

import (
	v1 "github.com/openshift-eng/ci-test-mapping/pkg/api/types/v1"
	"github.com/openshift-eng/ci-test-mapping/pkg/config"
)

type Component struct {
	*config.Component
}

var MonitoringComponent = Component{
	Component: &config.Component{
		Name:                 "Monitoring",
		Operators:            []string{"monitoring"},
		DefaultJiraComponent: "Monitoring",
		Matchers: []config.ComponentMatcher{
			{
				IncludeAll: []string{"bz-Monitoring"},
			},
			{
				IncludeAll: []string{"bz-monitoring"},
			},
			{
				SIG:      "sig-instrumentation",
				Priority: 1,
			},
			{
				IncludeAll: []string{"cluster-monitoring-operator"},
			},
		},
	},
}

// renamedTests is a map that maps new test names to the earliest name of the test. This lets us have a stable test ID
// across all releases.
var renamedTests = map[string]string{
	"[sig-instrumentation][Late] Alerts shouldn't report any alerts in firing or pending state apart from Watchdog and AlertmanagerReceiversNotConfigured and have no gaps in Watchdog firing [Skipped:Disconnected] [Suite:openshift/conformance/parallel]": "[sig-instrumentation][Late] Alerts shouldn't report any alerts in firing state apart from Watchdog and AlertmanagerReceiversNotConfigured [Suite:openshift/conformance/parallel]",
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
	// Look up the stable name for our test in our renamed tests map.
	if stableName, ok := renamedTests[test.Name]; ok {
		return stableName
	}

	return test.Name
}

func (c *Component) JiraComponents() (components []string) {
	components = []string{c.DefaultJiraComponent}
	for _, m := range c.Matchers {
		components = append(components, m.JiraComponent)
	}

	return components
}
