package terraform_test

import (
	"testing"

	"github.com/massdriver-cloud/schema-generator/pkg/terraform"
)

func TestRun(t *testing.T) {
	type testData struct {
		name       string
		modulePath string
		want       string
	}
	tests := []testData{
		{
			name:       "simple",
			modulePath: "testdata/simple/",
			want:       "foo",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := terraform.Run(tc.modulePath)
			if err != nil {
				t.Fatalf("%d, unexpected error", err)
			}
		})
	}
}
