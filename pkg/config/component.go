package config

import (
	"strings"

	v1 "github.com/openshift-eng/ci-test-mapping/pkg/api/types/v1"
	"github.com/openshift-eng/ci-test-mapping/pkg/util"
)

// Component is the default configuration struct that you can include in your
// own component implementation. It includes a matcher help that will identify
// if a test belongs to a sig, operator, as well as simple substring matching.
// Components do not need to use this framework, it's an optional add-on.
type Component struct {
	Name                 string
	DefaultJiraComponent string
	Matchers             []ComponentMatcher
	Operators            []string
}

// ComponentMatcher is used to match against a TestInfo struct. Note the fields SIG,
// Suite, Include, and Exclude are ANDed together. That is, all that have values must
// match.  For include  and exclude, the individual items in the array are ANDed. That
// is, if you  specify multiple substrings, all must match. Use separate component
// matchers for an OR operation.
//
// The second set  of fields are metadata used to assign ownership.
type ComponentMatcher struct {
	SIG     string
	Suite   string
	Include []string
	Exclude []string

	JiraComponent string
	Capabilities  []string
	Priority      int
}

func (c *Component) FindMatch(test *v1.TestInfo) *ComponentMatcher {
	if ok, capabilities := c.IsOperatorTest(test); ok {
		return &ComponentMatcher{
			JiraComponent: c.DefaultJiraComponent,
			Capabilities:  capabilities,
		}
	}

	// Check if any of the Matchers match the given test
	for _, m := range c.Matchers {
		sigMatch := true
		suiteMatch := true
		incSubstrMatch := true
		excSubstrMatch := true

		if m.SIG != "" {
			sigMatch = util.IsSigTest(test.Name, m.SIG)
		}

		if m.Suite != "" {
			suiteMatch = m.IsSuiteTest(test)
		}

		if len(m.Include) > 0 {
			incSubstrMatch = m.IsSubstringTest(test)
		}

		if len(m.Exclude) > 0 {
			excSubstrMatch = !m.IsSubstringTest(test)
		}

		// AND the three match results together
		if sigMatch && suiteMatch && incSubstrMatch && excSubstrMatch {
			return &m
		}
	}

	return nil
}

func (cm *ComponentMatcher) IsSuiteTest(test *v1.TestInfo) bool {
	return test.Suite == cm.Suite
}

func (cm *ComponentMatcher) IsSubstringTest(test *v1.TestInfo) bool {
	for _, str := range cm.Include {
		if !strings.Contains(test.Name, str) {
			return false
		}
	}
	return true
}

func (c *Component) IsOperatorTest(test *v1.TestInfo) (bool, []string) {
	for _, operator := range c.Operators {
		// OpenShift tests related to operators (install, upgrade, etc)
		if isOperatorTest, capabilities := util.IdentifyOperatorTest(operator, test.Name); isOperatorTest {
			return true, capabilities
		}
	}

	return false, nil
}
