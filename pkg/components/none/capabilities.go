package none

import (
	"strings"

	v1 "github.com/openshift-eng/ci-test-mapping/pkg/api/types/v1"
	"github.com/openshift-eng/ci-test-mapping/pkg/util"
)

func identifyCapabilities(test *v1.TestInfo) []string {
	var capabilities []string

	// Get the Feature name from the test name as a capability
	capabilities = append(capabilities, util.ExtractTestField(test.Name, "Feature")...)

	if strings.Contains(test.Name, "clusteroperator/") {
		capabilities = append(capabilities, "Operator")
	}

	if strings.Contains(test.Name, "alert/") {
		capabilities = append(capabilities, "Alerts")
	}

	return capabilities
}
