package util

import (
	"reflect"
	"testing"
)

func TestIdentifyOperatorTest(t *testing.T) {
	tests := []struct {
		name             string
		testName         string
		operator         string
		isOperatorTest   bool
		wantCapabilities []string
	}{
		{
			name:             "cluster upgrade match",
			testName:         "Cluster upgrade.Operator upgrade etcd",
			operator:         "etcd",
			isOperatorTest:   true,
			wantCapabilities: []string{"upgrade"},
		},
		{
			name:             "cluster upgrade non-match",
			testName:         "Cluster upgrade.Operator upgrade insights",
			operator:         "etcd",
			isOperatorTest:   false,
			wantCapabilities: nil,
		},
		{
			name:             "cluster upgrade other test",
			testName:         "Some other test",
			operator:         "etcd",
			isOperatorTest:   false,
			wantCapabilities: nil,
		},
		{
			name:             "cluster install",
			testName:         "operator install insights",
			operator:         "insights",
			isOperatorTest:   true,
			wantCapabilities: []string{"install"},
		},
		{
			name:             "cluster install non-match",
			testName:         "operator install insights",
			operator:         "etcd",
			isOperatorTest:   false,
			wantCapabilities: nil,
		},
		{
			name:             "cluster install other test",
			testName:         "Some other test",
			operator:         "etcd",
			isOperatorTest:   false,
			wantCapabilities: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOperatorTest, gotCapabilities := IdentifyOperatorTest(tt.operator, tt.testName)
			if gotOperatorTest != tt.isOperatorTest {
				t.Errorf("IdentifyOperatorTest() isOperatorTest = %v, want %v", gotOperatorTest, tt.isOperatorTest)
			}
			if !reflect.DeepEqual(gotCapabilities, tt.wantCapabilities) {
				t.Errorf("IdentifyOperatorTest() gotCapabilities = %v, want %v", gotCapabilities, tt.wantCapabilities)
			}
		})
	}
}
