package components

import (
	"reflect"
	"testing"

	v1 "github.com/openshift-eng/ci-test-mapping/pkg/api/types/v1"
	"github.com/openshift-eng/ci-test-mapping/pkg/components/lpacm"
	"github.com/openshift-eng/ci-test-mapping/pkg/components/lpcnv"
	"github.com/openshift-eng/ci-test-mapping/pkg/config"
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
		{"lp-interop--ACM--ACMVirt--GRC", "LP--ACM", "ACM+Virt"},
		{"lp-interop--CNV--ACMVirt--Testing VM workload.", "LP--CNV", "ACM+Virt"},
		{"lp-interop--ACM--ACM+Virtual--GRC", "LP--ACM", "ACM+Virt"},
		{"lp-interop--CNV--ACM+Virtual--Testing VM workload.", "LP--CNV", "ACM+Virt"},
		{"lp-interop--ACM--ACM+Virt", "LP--ACM", "ACM+Virt"},
		{"lp-interop--CNV--ACM+Virt", "LP--CNV", "ACM+Virt"},
		{"lp-interop--ACM--ACM+Virt-GRC", "LP--ACM", "ACM+Virt"},
		{"lp-interop--CNV--ACM+Virt-Testing VM workload.", "LP--CNV", "ACM+Virt"},
		{"lp-interop--ACM--standalone", "LP--ACM", "ACM+Virt"},
		{"lp-interop--CNV--standalone", "LP--CNV", "ACM+Virt"},
		{"lp-chaos--ACM--ACM+Virt--GRC", "LP--ACM", "ACM+Virt"},
		{"lp-chaos--CNV--ACM+Virt--Testing VM workload.", "LP--CNV", "ACM+Virt"},
		{"lp-ocp-compat--ACM--ACM+Virt--GRC", "LP--ACM", "ACM+Virt"},
		{"lp-ocp-compat--CNV--ACM+Virt--Testing VM workload.", "LP--CNV", "ACM+Virt"},
		{"lp-interop--OPP--Observability", "LP--ACM", "opp"},
		{"acm-opp-app", "LP--ACM", "opp"},
		{"lp-interop--OPP--acm", "LP--ACM", "opp"},
		{"lp-interop--ACM-OBS-ODF", "LP--ACM", "opp"},
		{"ACM-lp-interop", "LP--ACM", ""},
		{"CNV-lp-interop", "LP--CNV", "ACM+Virt"},
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

func TestLPMatcherPrecedence(t *testing.T) {
	t.Parallel()
	components := []struct {
		name      string
		component *config.Component
	}{
		{"ACM", lpacm.LPacmComponent.Component},
		{"CNV", lpcnv.LPcnvComponent.Component},
	}
	for _, component := range components {
		prefix := "lp-interop--" + component.name + "--"
		tests := []struct {
			suffix       string
			wantSpecific bool
		}{
			{"ACM+Virt--GRC", true},
			{"ACM+Virt--Testing VM workload.", true},
			{"ACMVirt--GRC", false},
			{"ACMMVirt--GRC", false},
			{"ACM+Virtual--GRC", false},
			{"ACM+Virt", false},
			{"ACM+Virt-GRC", false},
			{"standalone", false},
		}
		for _, testCase := range tests {
			t.Run(component.name+"/"+testCase.suffix, func(t *testing.T) {
				t.Parallel()
				wantPattern := "^" + prefix
				if testCase.wantSpecific {
					wantPattern += `ACM\+Virt--`
				}
				matcher := component.component.FindMatch(&v1.TestInfo{Suite: prefix + testCase.suffix})
				if matcher == nil || matcher.SuiteRegEx == nil {
					t.Fatalf("FindMatch() = %+v, want regex %q", matcher, wantPattern)
				}
				if matcher.SuiteRegEx.String() != wantPattern {
					t.Errorf("FindMatch() regex = %q, want %q", matcher.SuiteRegEx, wantPattern)
				}
			})
		}
		t.Run(component.name+"/anchored prefix", func(t *testing.T) {
			t.Parallel()
			if matcher := component.component.FindMatch(&v1.TestInfo{Suite: "unrelated--" + prefix + "ACM+Virt--GRC"}); matcher != nil {
				t.Errorf("FindMatch() = %+v, want no match", matcher)
			}
		})
		t.Run(component.name+"/exact flat suite", func(t *testing.T) {
			t.Parallel()
			suite := component.name + "-lp-interop"
			matcher := component.component.FindMatch(&v1.TestInfo{Suite: suite})
			if matcher == nil || matcher.Suite != suite || matcher.SuiteRegEx != nil {
				t.Errorf("FindMatch() = %+v, want exact suite %q", matcher, suite)
			}
			if matcher := component.component.FindMatch(&v1.TestInfo{Suite: suite + "--extra"}); matcher != nil {
				t.Errorf("FindMatch() = %+v, want no match for suffixed flat suite", matcher)
			}
		})
	}
}

func TestLPAdditionalCapabilities(t *testing.T) {
	t.Parallel()
	tests := []struct {
		suite            string
		name             string
		wantCapabilities []string
		wantComponent    string
	}{
		{"CNV-lp-interop", "[Feature:LPFeature] checks workload", []string{"ACM+Virt", "LPFeature"}, "LP--CNV"},
		{"CNV-lp-interop", "[Feature:ACM+Virt] checks workload", []string{"ACM+Virt"}, "LP--CNV"},
		{"ACM-lp-interop", "[Feature:LPFeature] checks workload", []string{"LPFeature"}, "LP--ACM"},
		{"lp-chaos--ACM--GRC", "[Feature:Chaos] checks workload", []string{"ACM+Virt", "Chaos"}, "LP--ACM"},
		{"lp-ocp-compat--CNV--Testing VM workload.", "[Feature:Compatibility] checks workload", []string{"ACM+Virt", "Compatibility"}, "LP--CNV"},
		{"lp-interop--OPP--Observability", "[Feature:LPFeature] checks workload", []string{"LPFeature", "opp"}, "LP--ACM"},
		{"CNV-lp-interop", "[Feature:ResourceQuota] checks quota", []string{"ResourceQuota"}, "kube-controller-manager"},
		{"lp-interop--OPP--Observability", "[Feature:ResourceQuota] checks quota", []string{"ResourceQuota"}, "kube-controller-manager"},
	}
	identifier := NewTestIdentifier(registry.NewComponentRegistry(), nil)
	for _, testCase := range tests {
		t.Run(testCase.suite+"/"+testCase.name, func(t *testing.T) {
			t.Parallel()
			testInfo := &v1.TestInfo{Name: testCase.name, Suite: testCase.suite}
			ownership, err := identifier.Identify(testInfo)
			if err != nil || ownership == nil {
				t.Fatalf("Identify() = %v, %v", ownership, err)
			}
			if !reflect.DeepEqual(ownership.Capabilities, testCase.wantCapabilities) {
				t.Errorf("Identify() capabilities = %v, want %v", ownership.Capabilities, testCase.wantCapabilities)
			}
			if ownership.Component != testCase.wantComponent {
				t.Errorf("Identify() component = %q, want %q", ownership.Component, testCase.wantComponent)
			}
			if ownership.Name != testInfo.Name || ownership.ID != util.StableID(testInfo, testInfo.Name) {
				t.Errorf("Identify() changed test identity: %+v", ownership)
			}
		})
	}
}
