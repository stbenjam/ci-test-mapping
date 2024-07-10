package kubeapiserver

import (
	"strings"

	v1 "github.com/openshift-eng/ci-test-mapping/pkg/api/types/v1"
	"github.com/openshift-eng/ci-test-mapping/pkg/util"
)

func identifyCapabilities(test *v1.TestInfo) []string {
	capabilities := util.DefaultCapabilities(test)

	if strings.Contains(test.Name, "disruption/") {
		capabilities = append(capabilities, "Disruption")
	}

	return capabilities
}
