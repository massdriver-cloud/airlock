package bicep_test

import (
	"testing"

	"github.com/massdriver-cloud/airlock/pkg/bicep"
	"github.com/stretchr/testify/require"
)

func TestBicepToSchema(t *testing.T) {
	type testData struct {
		name       string
		modulePath string
		want       string
	}
	tests := []testData{
		{
			name:       "simple",
			modulePath: "testdata/template.bicep",
			want: `
{
	"required": [
		"testArray",
		"testBool",
		"testInt",
		"testObject",
		"testString"
	],
	"type": "object",
	"properties": {
		"testString": {
			"title": "testString",
			"type": "string",
			"description": "an example string parameter",
			"default": "foo"
		},
		"testInt": {
			"title": "testInt",
			"type": "integer",
			"default": 1
		},
		"testBool": {
			"title": "testBool",
			"type": "boolean",
			"default": false
		},
		"testArray": {
			"title": "testArray",
			"type": "array",
			"default": [1, 2, 3]
		},
		"testObject": {
			"title": "testObject",
			"type": "object"
		}
	}
}
`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := bicep.BicepToSchema(tc.modulePath)
			if err != nil {
				t.Fatalf("%d, unexpected error", err)
			}
			require.JSONEq(t, tc.want, got)
		})
	}
}
