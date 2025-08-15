package helm_test

import (
	"encoding/json"
	"testing"

	"github.com/massdriver-cloud/airlock/pkg/helm"
	"github.com/massdriver-cloud/airlock/pkg/result"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	type testData struct {
		name       string
		valuesPath string
		want       string
		diags      []result.Diagnostic
	}
	tests := []testData{
		{
			name:       "simple",
			valuesPath: "testdata/values.yaml",
			want: `
{
	"required": [
		"hello",
		"age",
		"height",
		"object",
		"array",
		"emptyArray",
		"nullValue"
	],
	"type": "object",
	"properties": {
		"hello": {
			"title": "hello",
			"type": "string",
			"description": "An example string variable",
			"default": "world"
		},
		"age": {
			"title": "age",
			"type": "integer",
			"description": "An example integer variable",
			"default": 14
		},
		"height": {
			"title": "height",
			"type": "number",
			"description": "An example float variable",
			"default": 3.3
		},
		"object": {
			"title": "object",
			"type": "object",
			"properties": {
				"nestedString": {
					"title": "nestedString",
					"type": "string",
					"description": "A nested variable",
					"default": "a string"
				},
				"nestedBool": {
					"title": "nestedBool",
					"type": "boolean",
					"default": true
				}
			},
			"required": [
				"nestedString",
				"nestedBool"
			],
			"description": "An example object variable"
		},
		"array": {
			"title": "array",
			"type": "array",
			"description": "An example array variable",
			"items": {
				"type": "string",
				"default": "foo"
			},
			"default": [
				"foo",
				"bar"
			]
		},
		"emptyArray": {
			"title": "emptyArray",
			"type": "array",
			"description": "An empty array should not cause an error",
			"items": {
				"$comment": "Airlock Warning: unknown type from empty array"
			}
		},
		"nullValue": {
			"title": "nullValue",
			"$comment": "Airlock Warning: unknown type from null value"
		}
	}
}
`,
			diags: []result.Diagnostic{
				{
					Path:    "emptyArray",
					Code:    "unknown_type",
					Message: "array emptyArray is empty so it's type is unknown",
					Level:   result.Warning,
				},
				{
					Path:    "nullValue",
					Code:    "unknown_type",
					Message: "type of field nullValue is indeterminate (null)",
					Level:   result.Warning,
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := helm.HelmToSchema(tc.valuesPath)

			bytes, err := json.Marshal(got.Schema)
			if err != nil {
				t.Fatalf("%d, unexpected error", err)
			}

			require.JSONEq(t, tc.want, string(bytes))

			assert.ElementsMatch(t, tc.diags, got.Diags)
		})
	}
}
