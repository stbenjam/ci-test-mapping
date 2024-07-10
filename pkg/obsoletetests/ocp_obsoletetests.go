package obsoletetests

import (
	"k8s.io/apimachinery/pkg/util/sets"

	v1 "github.com/openshift-eng/ci-test-mapping/pkg/api/types/v1"
)

type OCPObsoleteTestManager struct{}

type obsoleteTestIdentifier struct {
	name  string
	suite string
}

var obsoleteTests = sets.New[obsoleteTestIdentifier](
	[]obsoleteTestIdentifier{
		// Removed in alert refactor by TRT https://github.com/openshift/origin/pull/28332
		{
			name:  "[sig-arch] Check if alerts are firing during or after upgrade success",
			suite: "Cluster upgrade",
		},
	}...)

func (*OCPObsoleteTestManager) IsObsolete(test *v1.TestInfo) bool {
	return obsoleteTests.Has(obsoleteTestIdentifier{name: test.Name, suite: test.Suite})
}
