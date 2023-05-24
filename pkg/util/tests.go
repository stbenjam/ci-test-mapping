package util

import (
	"fmt"
	"regexp"
	"strings"

	v1 "github.com/openshift-eng/ci-test-mapping/pkg/api/types/v1"
)

var fieldRegexp = regexp.MustCompile(`\[([^\]]*):([^\]]*)\]`)

// ExtractTestField gets the value of a field in a test name, e.g. [sig-storage][Driver: gce] would return
// "gce" when extracting "Driver"
func ExtractTestField(testName, field string) (results []string) {
	matches := fieldRegexp.FindAllStringSubmatch(testName, -1)
	for _, match := range matches {
		if len(match) == 3 && match[1] == field {
			results = append(results, strings.TrimSpace(match[2]))
		}
	}

	return results
}

// StableID produces a stable test ID based on a TestInfo struct.
func StableID(testInfo *v1.TestInfo) string {
	testName := testInfo.Name

	if testInfo.Suite != "" {
		return fmt.Sprintf("%s.%s", testInfo.Suite, testName)
	}

	return testName
}
