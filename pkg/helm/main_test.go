package helm_test

import (
	"reflect"
	"testing"

	"github.com/massdriver-cloud/schema-generator/pkg/helm"
)

func TestRun(t *testing.T) {
	type testData struct {
		name       string
		valuesPath string
		want       string
	}
	tests := []testData{
		{
			name:       "simple",
			valuesPath: "testdata/values.yaml",
			want:       "foo",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := helm.Run(tc.valuesPath)
			if err != nil {
				t.Fatalf("%d, unexpected error", err)
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}
