package components

import (
	"crypto/md5"
	"fmt"

	log "github.com/sirupsen/logrus"

	v1 "github.com/openshift-eng/ci-test-mapping/pkg/api/types/v1"
	"github.com/openshift-eng/ci-test-mapping/pkg/registry"
	"github.com/openshift-eng/ci-test-mapping/pkg/util"
)

const (
	DefaultComponent  = "Unknown"
	DefaultCapability = "Other"
	DefaultProduct    = "OpenShift"
)

func IdentifyTest(reg *registry.Registry, test *v1.TestInfo) (*v1.TestOwnership, error) {
	var ownerships []*v1.TestOwnership

	log.WithFields(testInfoLogFields(test)).Debugf("attempting to identify test using %d components", len(reg.Components))
	for name, component := range reg.Components {
		log.WithFields(testInfoLogFields(test)).Tracef("checking component %q", name)
		ownership, err := component.IdentifyTest(test)
		if err != nil {
			log.WithError(err).Errorf("component %q returned an error", name)
			return nil, err
		}
		if ownership != nil {
			log.WithFields(testInfoLogFields(test)).Tracef("component %q claimed this test", name)
			ownerships = append(ownerships, setDefaults(test, ownership, component))
		}
	}

	if len(ownerships) == 0 {
		ownerships = append(ownerships, setDefaults(test, &v1.TestOwnership{
			ID:   fmt.Sprintf("%x", md5.Sum([]byte(util.StableID(test)))),
			Name: test.Name,
		}, nil))
	}

	return getHighestPriority(ownerships)
}

func setDefaults(testInfo *v1.TestInfo, testOwnership *v1.TestOwnership, c v1.Component) *v1.TestOwnership {
	if testOwnership.ID == "" && c != nil {
		testOwnership.ID = fmt.Sprintf("%x", md5.Sum([]byte(c.StableID(testInfo))))
	}

	testOwnership.Kind = v1.Kind
	testOwnership.APIVersion = v1.APIVersion

	if testOwnership.Product == "" {
		testOwnership.Product = DefaultProduct
	}

	if testOwnership.Component == "" {
		testOwnership.Component = DefaultComponent
	}

	if len(testOwnership.Capabilities) == 0 {
		testOwnership.Capabilities = []string{DefaultCapability}
	}

	if testOwnership.Suite == "" {
		testOwnership.Suite = testInfo.Suite
	}

	return testOwnership
}

func testInfoLogFields(testInfo *v1.TestInfo) log.Fields {
	return log.Fields{
		"name":  testInfo.Name,
		"suite": testInfo.Suite,
	}
}

func getHighestPriority(ownerships []*v1.TestOwnership) (*v1.TestOwnership, error) {
	var highest *v1.TestOwnership
	for _, ownership := range ownerships {
		if highest != nil && ownership.Priority == highest.Priority {
			return nil, fmt.Errorf("test %q is claimed by %s, %s - unable to resolve conflict "+
				"-- please use priority field", highest.Name, highest.Component, ownership.Component)
		}

		if highest == nil || ownership.Priority > highest.Priority {
			highest = ownership
		}
	}

	return highest, nil
}
