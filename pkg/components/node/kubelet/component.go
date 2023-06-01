package nodekubelet

import (
	v1 "github.com/openshift-eng/ci-test-mapping/pkg/api/types/v1"
	"github.com/openshift-eng/ci-test-mapping/pkg/config"
)

type Component struct {
	*config.Component
}

var KubeletComponent = Component{
	Component: &config.Component{
		Name:                 "Node / Kubelet",
		Operators:            []string{},
		DefaultJiraComponent: "Node / Kubelet",
		Matchers: []config.ComponentMatcher{
			{
				SIG:      "sig-node",
				Priority: 1,
			},
			{
				IncludeAny: []string{
					"Node process segfaulted",
					"all nodes should be ready",
					"[sig-arch] [Conformance] sysctl pod should not start for sysctl not on whitelist kernel.msgmax [Suite:openshift/conformance/parallel/minimal]",
					"[sig-arch] [Conformance] sysctl pod should not start for sysctl not on whitelist net.ipv4.ip_dynaddr [Suite:openshift/conformance/parallel/minimal]",
					"[sig-arch] [Conformance] sysctl whitelists kernel.shm_rmid_forced [Suite:openshift/conformance/parallel/minimal]",
					"[sig-arch] [Conformance] sysctl whitelists net.ipv4.ip_local_port_range [Suite:openshift/conformance/parallel/minimal]",
					"[sig-arch] [Conformance] sysctl whitelists net.ipv4.ip_unprivileged_port_start [Suite:openshift/conformance/parallel/minimal]",
					"[sig-arch] [Conformance] sysctl whitelists net.ipv4.ping_group_range [Suite:openshift/conformance/parallel/minimal]",
					"[sig-arch] [Conformance] sysctl whitelists net.ipv4.tcp_syncookies [Suite:openshift/conformance/parallel/minimal]",
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
	// Look up the stable name for our test in our renamed tests map.
	if stableName, ok := c.TestRenames[test.Name]; ok {
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
