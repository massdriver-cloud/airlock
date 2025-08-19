package opentofu_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/massdriver-cloud/airlock/pkg/opentofu"
	"github.com/massdriver-cloud/airlock/pkg/result"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTofuToSchema(t *testing.T) {
	type testData struct {
		name  string
		diags []result.Diagnostic
	}
	tests := []testData{
		{
			name: "simple",
			diags: []result.Diagnostic{
				{
					Path:    "any",
					Code:    "unconstrained_type",
					Message: "unconstrained type in field 'any' from OpenTofu/Terraform 'any'",
					Level:   result.Warning,
				},
				{
					Path:    "foo",
					Code:    "unconstrained_type",
					Message: "unconstrained type in field 'foo' from OpenTofu/Terraform 'any'",
					Level:   result.Warning,
				},
				{
					Path:    "empty",
					Code:    "unconstrained_type",
					Message: "unconstrained type in field 'empty' from OpenTofu/Terraform 'any'",
					Level:   result.Warning,
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			modulePath := filepath.Join("testdata/opentofu", tc.name)

			got := opentofu.TofuToSchema(modulePath)

			gotSchema, marshalErr := json.Marshal(got.Schema)
			if marshalErr != nil {
				t.Fatalf("unexpected error: %s", marshalErr.Error())
			}

			wantSchema, readErr := os.ReadFile(filepath.Join("testdata/opentofu", tc.name, "schema.json"))
			if readErr != nil {
				t.Fatalf("unexpected error: %s", readErr.Error())
			}

			require.JSONEq(t, string(wantSchema), string(gotSchema))

			gotDiags := got.Diags

			assert.ElementsMatch(t, tc.diags, gotDiags)
		})
	}
}
