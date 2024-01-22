package kubecontrollermanager

import (
	"strings"

	v1 "github.com/openshift-eng/ci-test-mapping/pkg/api/types/v1"
	"github.com/openshift-eng/ci-test-mapping/pkg/util"
)

func identifyCapabilities(test *v1.TestInfo) []string {
	capabilities := util.DefaultCapabilities(test)

	if strings.Contains(test.Name, "ClusterResourceQuota") {
		capabilities = append(capabilities, "ClusterResourceQuota")
	}

	if strings.Contains(test.Name, "ResourceQuota") {
		capabilities = append(capabilities, "ResourceQuota")
	}

	return capabilities
}
