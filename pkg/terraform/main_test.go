package terraform_test

import (
	"testing"

	"github.com/massdriver-cloud/airlock/pkg/terraform"

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
			modulePath: "testdata/simple/",
			want: `
{
	"required": [
		"teststring",
		"testnumber",
		"testbool",
		"testobject",
		"testlist",
		"testset",
		"testmap",
		"nodescription"
	],
	"properties": {
		"teststring": {
			"title": "teststring",
			"type": "string",
			"description": "An example string variable",
			"default": "string value"
		},
		"testnumber": {
			"title": "testnumber",
			"type": "number",
			"description": "An example number variable",
			"default": 20
		},
		"testbool": {
			"title": "testbool",
			"type": "boolean",
			"description": "An example bool variable",
			"default": false
		},
		"testobject": {
			"title": "testobject",
			"type": "object",
			"properties": {
				"name": {
					"title": "name",
					"type": "string"
				},
				"address": {
					"title": "address",
					"type": "string"
				},
				"age": {
					"title": "age",
					"type": "number"
				}
			},
			"required": [
				"name",
				"address"
			],
			"description": "An example object variable",
			"default": {
				"name": "Bob",
				"address": "123 Bob St."
			}
		},
		"testlist": {
			"title": "testlist",
			"type": "array",
			"description": "An example list variable",
			"items": {
				"type": "string"
			}
		},
		"testset": {
			"title": "testset",
			"type": "array",
			"uniqueItems": true,
			"description": "An example set variable",
			"items": {
				"type": "string"
			}
		},
		"testmap": {
			"title": "testmap",
			"type": "object",
			"description": "An example map variable",
			"propertyNames": {
				"pattern": "^.*$"
			},
			"additionalProperties": {
				"type": "string"
			}
		},
		"nodescription": {
			"title": "nodescription",
			"type": "string"
		}
	}
}
`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := terraform.Run(tc.modulePath)
			if err != nil {
				t.Fatalf("%d, unexpected error", err)
			}
			require.JSONEq(t, got, tc.want)
		})
	}
}
