package kubeapiserver

import (
	v1 "github.com/openshift-eng/ci-test-mapping/pkg/api/types/v1"
	"github.com/openshift-eng/ci-test-mapping/pkg/config"
)

type Component struct {
	*config.Component
}

var KubeApiserverComponent = Component{
	Component: &config.Component{
		Name:                 "kube-apiserver",
		Operators:            []string{"kube-apiserver"},
		DefaultJiraComponent: "kube-apiserver",
		Namespaces: []string{
			"default",
			"openshift",
			"kube-system",
			"openshift-config",
			"openshift-config-managed",
			"openshift-kube-apiserver",
			"openshift-kube-apiserver-operator",
		},
		Matchers: []config.ComponentMatcher{
			{
				IncludeAny: []string{
					"bz-kube-apiserver",
					"cache-kube-api-",
					"[sig-api-machinery][Feature:APIServer]",
					"should have a status in the CRD schema", // only observed failures are kube-apiserver availability.
				},
			},
			{
				IncludeAll: []string{"kube-api-", "-connections"},
			},
			{
				SIG:      "sig-api-machinery",
				Priority: -1,
			},
		},
		TestRenames: map[string]string{
			"[Unknown][invariant] alert/KubePodNotReady should not be at or above info in ns/default":                                     "[bz-Unknown][invariant] alert/KubePodNotReady should not be at or above info in ns/default",
			"[Unknown][invariant] alert/KubePodNotReady should not be at or above info in ns/kube-system":                                 "[bz-Unknown][invariant] alert/KubePodNotReady should not be at or above info in ns/kube-system",
			"[Unknown][invariant] alert/KubePodNotReady should not be at or above info in ns/openshift":                                   "[bz-Unknown][invariant] alert/KubePodNotReady should not be at or above info in ns/openshift",
			"[Unknown][invariant] alert/KubePodNotReady should not be at or above info in ns/openshift-config":                            "[bz-Unknown][invariant] alert/KubePodNotReady should not be at or above info in ns/openshift-config",
			"[Unknown][invariant] alert/KubePodNotReady should not be at or above info in ns/openshift-config-managed":                    "[bz-Unknown][invariant] alert/KubePodNotReady should not be at or above info in ns/openshift-config-managed",
			"[Unknown][invariant] alert/KubePodNotReady should not be at or above pending in ns/default":                                  "[bz-Unknown][invariant] alert/KubePodNotReady should not be at or above pending in ns/default",
			"[Unknown][invariant] alert/KubePodNotReady should not be at or above pending in ns/kube-system":                              "[bz-Unknown][invariant] alert/KubePodNotReady should not be at or above pending in ns/kube-system",
			"[Unknown][invariant] alert/KubePodNotReady should not be at or above pending in ns/openshift":                                "[bz-Unknown][invariant] alert/KubePodNotReady should not be at or above pending in ns/openshift",
			"[Unknown][invariant] alert/KubePodNotReady should not be at or above pending in ns/openshift-config":                         "[bz-Unknown][invariant] alert/KubePodNotReady should not be at or above pending in ns/openshift-config",
			"[Unknown][invariant] alert/KubePodNotReady should not be at or above pending in ns/openshift-config-managed":                 "[bz-Unknown][invariant] alert/KubePodNotReady should not be at or above pending in ns/openshift-config-managed",
			"[kube-apiserver][invariant] alert/KubePodNotReady should not be at or above info in ns/openshift-kube-apiserver":             "[bz-kube-apiserver][invariant] alert/KubePodNotReady should not be at or above info in ns/openshift-kube-apiserver",
			"[kube-apiserver][invariant] alert/KubePodNotReady should not be at or above info in ns/openshift-kube-apiserver-operator":    "[bz-kube-apiserver][invariant] alert/KubePodNotReady should not be at or above info in ns/openshift-kube-apiserver-operator",
			"[kube-apiserver][invariant] alert/KubePodNotReady should not be at or above pending in ns/openshift-kube-apiserver":          "[bz-kube-apiserver][invariant] alert/KubePodNotReady should not be at or above pending in ns/openshift-kube-apiserver",
			"[kube-apiserver][invariant] alert/KubePodNotReady should not be at or above pending in ns/openshift-kube-apiserver-operator": "[bz-kube-apiserver][invariant] alert/KubePodNotReady should not be at or above pending in ns/openshift-kube-apiserver-operator",
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
