package helm_test

import (
	"testing"

	"github.com/massdriver-cloud/airlock/pkg/helm"
	"github.com/stretchr/testify/require"
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
			modulePath: "testdata/values.yaml",
			want: `
{
	"required": [
		"hello",
		"age",
		"height",
		"object",
		"array"
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
		}
	}
}
`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := helm.Run(tc.modulePath)
			if err != nil {
				t.Fatalf("%d, unexpected error", err)
			}
			require.JSONEq(t, got, tc.want)
		})
	}
}
