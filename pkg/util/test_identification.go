package util

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	conditions   = regexp.MustCompile(`operator conditions (.*)`)
	upgradeRegex = regexp.MustCompile(`Operator upgrade (.*)`)
	installRegex = regexp.MustCompile("operator install (.*)")
	imageBuild   = regexp.MustCompile("Build image (.*) from the repository")
)

func IsSigTest(testName, sigName string) bool {
	return strings.Contains(testName, fmt.Sprintf("[%s]", sigName))
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
