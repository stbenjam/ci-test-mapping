package clusterversionoperator

import (
	v1 "github.com/openshift-eng/ci-test-mapping/pkg/api/types/v1"
	"github.com/openshift-eng/ci-test-mapping/pkg/config"
	"github.com/openshift-eng/ci-test-mapping/pkg/util"
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
				IncludeAny: []string{
					"cluster-version-operator",
					"bz-Cluster Version Operator",
				},
			},
			{
				IncludeAll: []string{"upgrade"},
				// Let others claim upgrade tests (i.e. for their component)
				Priority: -10,
			},
			{
				IncludeAny: []string{
					"[sig-arch] ClusterOperators [apigroup:config.openshift.io] should define at least one namespace in their lists of related objects  [Suite:openshift/conformance/parallel]",
					"[sig-arch] ClusterOperators [apigroup:config.openshift.io] should define at least one namespace in their lists of related objects [Suite:openshift/conformance/parallel]",
					"[sig-arch] ClusterOperators [apigroup:config.openshift.io] should define at least one related object that is not a namespace [Suite:openshift/conformance/parallel]",
					"[sig-arch] ClusterOperators [apigroup:config.openshift.io] should define valid related objects [Suite:openshift/conformance/parallel]",
					"[sig-arch] ClusterOperators should define at least one namespace in their lists of related objects [Suite:openshift/conformance/parallel]",
					"[sig-arch] ClusterOperators should define at least one related object that is not a namespace [Suite:openshift/conformance/parallel]",
					"[sig-arch] ClusterOperators should define valid related objects [Suite:openshift/conformance/parallel]",
					"[sig-cluster-lifecycle] TestAdminAck should succeed [apigroup:config.openshift.io] [Suite:openshift/conformance/parallel]",
					"[sig-cluster-lifecycle] TestAdminAck should succeed [Suite:openshift/conformance/parallel]",
				},
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
