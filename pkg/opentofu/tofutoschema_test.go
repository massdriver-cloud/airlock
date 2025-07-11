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
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			modulePath := filepath.Join("testdata/opentofu", tc.name)

			got, err := opentofu.TofuToSchema(modulePath)
			if err != nil && tc.err == "" {
				t.Fatalf("unexpected error: %s", err.Error())
			}
			if tc.err != "" && err == nil {
				t.Fatalf("expected error %s, got nil", tc.err)
			}
			if tc.err != "" && !strings.Contains(err.Error(), tc.err) {
				t.Fatalf("expected error %s, got %s", tc.err, err.Error())
			}
			if tc.err != "" {
				return
			}

			bytes, err := json.Marshal(got)
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}

			want, err := os.ReadFile(filepath.Join("testdata/opentofu", tc.name, "schema.json"))
			if err != nil {
				t.Fatalf("unexpected error: %s", err.Error())
			}

			require.JSONEq(t, string(want), string(bytes))
		})
	}
}
