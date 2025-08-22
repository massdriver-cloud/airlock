package bicep_test

import (
	"encoding/json"
	"testing"

	"github.com/massdriver-cloud/airlock/pkg/bicep"
	"github.com/massdriver-cloud/airlock/pkg/result"
	"github.com/stretchr/testify/assert"
)

func TestBicepToSchema(t *testing.T) {
	type testData struct {
		name      string
		bicepPath string
		diags     []result.Diagnostic
		want      string
	}
	tests := []testData{
		{
			name:      "simple",
			bicepPath: "testdata/template.bicep",
			diags:     []result.Diagnostic{},
			want: `
{
	"required": [
		"testArray",
		"testArrayObject",
		"testBool",
		"testEmptyArray",
		"testEmptyObject",
		"testInt",
		"testObject",
		"testSecureObject",
		"testSecureString",
		"testString"
	],
	"type": "object",
	"properties": {
		"testString": {
			"title": "testString",
			"type": "string",
			"description": "an example string parameter",
			"minLength": 2,
			"maxLength": 20,
			"enum": ["foo","bar"],
			"default": "foo"
		},
		"testInt": {
			"title": "testInt",
			"type": "integer",
			"minimum": 0,
			"maximum": 10,
			"enum": [1,5,7],
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
			"minItems": 1,
			"maxItems": 8,
			"default": [1, 2, 3],
			"items": {
				"type": "integer"
			}
		},
		"testObject": {
			"title": "testObject",
			"type": "object",
			"required": ["age","empty","friends","member","name","nested"],
			"properties": {
				"name": {
					"type": "string",
					"title": "name",
					"default": "hugh"
				},
				"age": {
					"type": "integer",
					"title": "age",
					"default": 20
				},
				"member": {
					"type": "boolean",
					"title": "member",
					"default": true
				},
				"nested": {
					"type": "object",
					"title": "nested",
					"required": ["foo","nested2"],
					"properties": {
						"foo": {
							"type": "string",
							"title": "foo",
							"default": "bar"
						},
						"nested2": {
							"type": "object",
							"title": "nested2",
							"required": ["hello"],
							"properties": {
								"hello": {
									"type": "string",
									"title": "hello",
									"default": "world"
								}
							}
						}
					}
				},
				"friends": {
					"type": "array",
					"title": "friends",
					"default": ["steve", "bob"],
					"items": {
						"type": "string"
					}
				},
				"empty": {
					"type": "array",
					"title": "empty"
				}
			}
		},
		"testArrayObject": {
			"type": "array",
			"title": "testArrayObject",
			"items": {
				"type": "object",
				"required": ["foo", "num"],
				"properties": {
					"foo": {
						"type": "string",
						"title": "foo",
						"default": "bar"
					},
					"num": {
						"type": "integer",
						"title": "num",
						"default": 10
					}
				}
			}
		},
		"testEmptyArray": {
			"type": "array",
			"title": "testEmptyArray"
		},
		"testEmptyObject": {
			"type": "object",
			"title": "testEmptyObject"
		},
		"testSecureObject": {
			"type": "object",
			"title": "testSecureObject"
		},
		"testSecureString": {
			"type": "string",
			"title": "testSecureString",
			"format": "password"
		}
	}
}
`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := bicep.BicepToSchema(tc.bicepPath)

			bytes, err := json.Marshal(got.Schema)
			if err != nil {
				t.Fatalf("%d, unexpected error", err)
			}

			assert.ElementsMatch(t, tc.diags, got.Diags)

			assert.JSONEq(t, tc.want, string(bytes))
		})
	}
}
