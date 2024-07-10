package storage

import (
	v1 "github.com/openshift-eng/ci-test-mapping/pkg/api/types/v1"
	"github.com/openshift-eng/ci-test-mapping/pkg/util"
)

func identifyCapabilities(test *v1.TestInfo) []string {
	capabilities := util.DefaultCapabilities(test)

	// Storage tests use Testpattern
	capabilities = append(capabilities, util.ExtractTestField(test.Name, "Testpattern")...)

	return capabilities
}
