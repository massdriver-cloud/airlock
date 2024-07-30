package terraform_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/massdriver-cloud/airlock/pkg/terraform"

	"github.com/stretchr/testify/require"
)

func TestTfToSchema(t *testing.T) {
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
			modulePath := filepath.Join("testdata/terraform", tc.name)

			want, err := os.ReadFile(filepath.Join("testdata/terraform", tc.name, "schema.json"))
			if err != nil {
				t.Fatalf("%d, unexpected error", err)
			}

			got, err := terraform.TfToSchema(modulePath)
			if err != nil {
				t.Fatalf("%d, unexpected error", err)
			}

			bytes, err := json.Marshal(got)
			if err != nil {
				t.Fatalf("%d, unexpected error", err)
			}

			require.JSONEq(t, string(want), string(bytes))
		})
	}
}
