package util

import (
	"reflect"
	"testing"
)

func TestExtractField(t *testing.T) {
	tests := []struct {
		name       string
		test       string
		field      string
		wantValues []string
	}{
		{
			name:       "can extract bracketed single value",
			test:       "[sig-storage] In-tree Volumes [Driver: windows-gcepd] [Testpattern: Dynamic PV (ntfs)][Feature:Windows] subPath should be able to unmount after the subpath directory is deleted [LinuxOnly] [Skipped:NoOptionalCapabilities] [Suite:openshift/conformance/parallel] [Suite:k8s]",
			field:      "Driver",
			wantValues: []string{"windows-gcepd"},
		},
		{
			name:       "can extract slash single value",
			test:       `jira/"Test Framework" validate the thing works`,
			field:      "jira",
			wantValues: []string{"Test Framework"},
		},
		{
			name:       "can extract fields case-insensitive",
			test:       `jira/"Test Framework" [Jira: Networking]`,
			field:      "JIRA",
			wantValues: []string{"Test Framework", "Networking"},
		},
		{
			name:       "can extract slash multiple value",
			test:       `jira/"Test Framework" jira/Installer validate the thing works`,
			field:      "jira",
			wantValues: []string{"Test Framework", "Installer"},
		},
		{
			name:       "handles field not present",
			test:       "[sig-storage] Foobar",
			field:      "Driver",
			wantValues: nil,
		},
		{
			name:       "can extract multiple values",
			test:       "[sig-storage] [Driver: aws] [Driver: gcp]",
			field:      "Driver",
			wantValues: []string{"aws", "gcp"},
		},
		{
			name:       "values with whitespace",
			test:       "[sig-storage] In-tree Volumes [Driver: azure-disk] [Testpattern: Dynamic PV (default fs)] subPath should support readOnly file specified in the volumeMount [LinuxOnly] [Suite:openshift/conformance/parallel] [Suite:k8s]",
			field:      "Testpattern",
			wantValues: []string{"Dynamic PV (default fs)"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResults := ExtractTestField(tt.test, tt.field); !reflect.DeepEqual(gotResults, tt.wantValues) {
				t.Errorf("ExtractTestField() = %v, want %v", gotResults, tt.wantValues)
			}
		})
	}
}
