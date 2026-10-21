package components

import (
	"reflect"
	"testing"

	v1 "github.com/openshift-eng/ci-test-mapping/pkg/api/types/v1"
	"github.com/openshift-eng/ci-test-mapping/pkg/registry"
	"github.com/openshift-eng/ci-test-mapping/pkg/util"
)

func TestLPCapabilities(t *testing.T) {
	t.Parallel()
	tests := []struct {
		suite      string
		component  string
		capability string
	}{
		{"lp-interop--ACM--ACM+Virt--GRC", "LP--ACM", "ACM+Virt"},
		{"lp-interop--ACM--ACM+Virt--observability", "LP--ACM", "ACM+Virt"},
		{"lp-interop--CNV--ACM+Virt--Testing VM workload.", "LP--CNV", "ACM+Virt"},
		{"lp-interop--ACM-Virt--GRC", "LP--ACM", "ACM+Virt"},
		{"lp-interop--ACM-Virt--observability", "LP--ACM", "ACM+Virt"},
		{"lp-interop--ACM--ACMVirt--GRC", "LP--ACM", "acm-virt"},
		{"lp-interop--CNV--ACMVirt--Testing VM workload.", "LP--CNV", "acm-virt"},
		{"lp-interop--ACM--ACM+Virtual--GRC", "LP--ACM", "acm-virt"},
		{"lp-interop--CNV--ACM+Virtual--Testing VM workload.", "LP--CNV", "acm-virt"},
		{"lp-interop--ACM--ACM+Virt", "LP--ACM", "acm-virt"},
		{"lp-interop--CNV--ACM+Virt", "LP--CNV", "acm-virt"},
		{"lp-interop--ACM--ACM+Virt-GRC", "LP--ACM", "acm-virt"},
		{"lp-interop--CNV--ACM+Virt-Testing VM workload.", "LP--CNV", "acm-virt"},
		{"lp-interop--ACM--standalone", "LP--ACM", "acm-virt"},
		{"lp-interop--CNV--standalone", "LP--CNV", "acm-virt"},
		{"lp-chaos--ACM--ACM+Virt--GRC", "LP--ACM", "acm-virt"},
		{"lp-chaos--CNV--ACM+Virt--Testing VM workload.", "LP--CNV", "acm-virt"},
		{"lp-ocp-compat--ACM--ACM+Virt--GRC", "LP--ACM", "acm-virt"},
		{"lp-ocp-compat--CNV--ACM+Virt--Testing VM workload.", "LP--CNV", "acm-virt"},
		{"lp-interop--OPP--Observability", "LP--ACM", "opp"},
		{"acm-opp-app", "LP--ACM", "opp"},
		{"lp-interop--OPP--acm", "LP--ACM", "opp"},
		{"lp-interop--ACM-OBS-ODF", "LP--ACM", "opp"},
		{"ACM-lp-interop", "LP--ACM", ""},
		{"CNV-lp-interop", "LP--CNV", ""},
	}
	componentRegistry := registry.NewComponentRegistry()
	identifier := NewTestIdentifier(componentRegistry, nil)
	for _, testCase := range tests {
		t.Run(testCase.suite, func(t *testing.T) {
			t.Parallel()
			testInfo := &v1.TestInfo{Name: "should preserve VM workload availability", Suite: testCase.suite}
			component := componentRegistry.Components[testCase.component]
			ownership, err := component.IdentifyTest(testInfo)
			if err != nil || ownership == nil {
				t.Fatalf("IdentifyTest() = %v, %v", ownership, err)
			}
			var wantCapabilities []string
			if testCase.capability != "" {
				wantCapabilities = []string{testCase.capability}
			}
			if !reflect.DeepEqual(ownership.Capabilities, wantCapabilities) {
				t.Errorf("IdentifyTest() capabilities = %v, want %v", ownership.Capabilities, wantCapabilities)
			}
			if component.StableID(testInfo) != testInfo.Name {
				t.Errorf("StableID() changed test name %q", testInfo.Name)
			}
			ownership, err = identifier.Identify(testInfo)
			if err != nil || ownership == nil {
				t.Fatalf("Identify() = %v, %v", ownership, err)
			}
			if testCase.capability == "" {
				wantCapabilities = []string{DefaultCapability}
			}
			if !reflect.DeepEqual(ownership.Capabilities, wantCapabilities) {
				t.Errorf("Identify() capabilities = %v, want %v", ownership.Capabilities, wantCapabilities)
			}
			if ownership.Component != testCase.component || ownership.JIRAComponent != testCase.component || ownership.Priority != 0 {
				t.Errorf("Identify() ownership = %+v, want component/Jira %s and priority 0", ownership, testCase.component)
			}
			if ownership.Name != testInfo.Name || ownership.ID != util.StableID(testInfo, testInfo.Name) {
				t.Errorf("Identify() changed test identity: %+v", ownership)
			}
		})
	}
	for _, component := range []string{"ACM", "CNV"} {
		t.Run(component+" feature capability is unique", func(t *testing.T) {
			t.Parallel()
			testInfo := &v1.TestInfo{
				Name:  "[Feature:ACM+Virt] should preserve VM workload availability",
				Suite: "lp-interop--" + component + "--ACM+Virt--Testing VM workload.",
			}
			ownership, err := identifier.Identify(testInfo)
			if err != nil || ownership == nil {
				t.Fatalf("Identify() = %v, %v", ownership, err)
			}
			if !reflect.DeepEqual(ownership.Capabilities, []string{"ACM+Virt"}) {
				t.Errorf("Identify() capabilities = %v, want exactly [ACM+Virt]", ownership.Capabilities)
			}
		})
	}
}
