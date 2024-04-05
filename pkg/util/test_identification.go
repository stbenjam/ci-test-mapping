package util

import (
	"fmt"
	"regexp"
	"strings"

	v1 "github.com/openshift-eng/ci-test-mapping/pkg/api/types/v1"
)

var (
	conditions      = regexp.MustCompile(`operator conditions (.*)`)
	upgradeRegex    = regexp.MustCompile(`Operator upgrade (.*)`)
	installRegex    = regexp.MustCompile("operator install (.*)")
	imageBuild      = regexp.MustCompile("Build image (.*) from the repository")
	disruptionRegex = regexp.MustCompile("disruption/|connection.*should be available|remains available|single second disruptions")
)

func DefaultCapabilities(test *v1.TestInfo) []string {
	var capabilities []string

	// Get the Feature name from the test name as a capability
	capabilities = append(capabilities, ExtractTestField(test.Name, "Feature")...)

	for _, featureGate := range ExtractTestField(test.Name, "FeatureGate") {
		capabilities = append(capabilities, fmt.Sprintf("FeatureGate:%s", featureGate))
	}

	for _, featureGate := range ExtractTestField(test.Name, "OCPFeatureGate") {
		capabilities = append(capabilities, fmt.Sprintf("OCPFeatureGate:%s", featureGate))
	}

	if strings.Contains(test.Name, "clusteroperator/") {
		capabilities = append(capabilities, "Operator")
	}

	if strings.Contains(test.Name, "alert/") {
		capabilities = append(capabilities, "Alerts")
	}

	if IsDisruptionTest(test.Name) {
		capabilities = append(capabilities, "Disruption")
	}

	return capabilities
}

func IsSigTest(testName, sigName string) bool {
	return strings.Contains(testName, fmt.Sprintf("[%s]", sigName))
}

func IsDisruptionTest(testName string) bool {
	return disruptionRegex.MatchString(testName)
}

func IdentifyOperatorTest(operator, testName string) (isOperatorTest bool, capabilities []string) {
	if matchOne(conditions, testName, operator) {
		return true, []string{"operator-conditions"}
	}

	if matchOne(upgradeRegex, testName, operator) {
		return true, []string{"upgrade"}
	}

	if matchOne(installRegex, testName, operator) {
		return true, []string{"install"}
	}

	if matchOne(imageBuild, testName, operator) {
		return true, []string{"images"}
	}

	return false, nil
}

func matchOne(re *regexp.Regexp, testName, match string) bool {
	matches := re.FindStringSubmatch(testName)
	if len(matches) > 1 && matches[1] == match {
		return true
	}

	return false
}
