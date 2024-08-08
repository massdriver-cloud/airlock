package bicep_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/massdriver-cloud/airlock/pkg/bicep"
)

func TestSchemaToBicep(t *testing.T) {
	type testData struct {
		name string
	}
	tests := []testData{
		{
			name: "simple",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			want, err := os.ReadFile(filepath.Join("testdata", tc.name+".bicep"))
			if err != nil {
				t.Fatalf("%d, unexpected error", err)
			}

			schemaFile, err := os.Open(filepath.Join("testdata", tc.name+".json"))
			if err != nil {
				t.Fatalf("%d, unexpected error", err)
			}

			got, err := bicep.SchemaToBicep(schemaFile)
			if err != nil {
				t.Fatalf("%d, unexpected error", err)
			}

			if string(got) != string(want) {
				t.Fatalf("\ngot: %q\n want: %q", string(got), string(want))
			}
		})
	}
}
