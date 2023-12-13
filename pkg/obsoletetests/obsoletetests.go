package obsoletetests

import v1 "github.com/openshift-eng/ci-test-mapping/pkg/api/types/v1"

type ObsoleteTestManager interface {
	IsObsolete(*v1.TestInfo) bool
}
