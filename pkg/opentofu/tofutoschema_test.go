package opentofu_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/massdriver-cloud/airlock/pkg/opentofu"

	"github.com/stretchr/testify/require"
)

func TestTofuToSchema(t *testing.T) {
	type testData struct {
		name string
		err  string
	}
	tests := []testData{
		{
			name: "simple",
		},
		{
			name: "any",
			err:  "type 'any' cannot be converted to a JSON schema type",
		},
		{
			name: "nestedany",
			err:  "dynamic types are not supported (are you using type 'any'?)",
		},
		{
			name: "empty",
			err:  "type cannot be empty",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			modulePath := filepath.Join("testdata/opentofu", tc.name)

			got, schemaErr := opentofu.TofuToSchema(modulePath)
			if schemaErr != nil && tc.err == "" {
				t.Fatalf("unexpected error: %s", schemaErr.Error())
			}
			if tc.err != "" && schemaErr == nil {
				t.Fatalf("expected error %s, got nil", tc.err)
			}
			if tc.err != "" && !strings.Contains(schemaErr.Error(), tc.err) {
				t.Fatalf("expected error %s, got %s", tc.err, schemaErr.Error())
			}
			if tc.err != "" {
				return
			}

			bytes, marshalErr := json.Marshal(got)
			if marshalErr != nil {
				t.Fatalf("unexpected error: %s", marshalErr.Error())
			}

			want, readErr := os.ReadFile(filepath.Join("testdata/opentofu", tc.name, "schema.json"))
			if readErr != nil {
				t.Fatalf("unexpected error: %s", readErr.Error())
			}

			require.JSONEq(t, string(want), string(bytes))
		})
	}
}
