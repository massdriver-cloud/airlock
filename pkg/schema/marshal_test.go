package schema_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/massdriver-cloud/airlock/pkg/schema"
	"github.com/stretchr/testify/require"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

func TestMarshal(t *testing.T) {
	type testData struct {
		name   string
		schema schema.Schema
	}
	tests := []testData{
		{
			name: "schema",
			schema: schema.Schema{
				Properties: orderedmap.New[string, *schema.Schema](orderedmap.WithInitialData[string, *schema.Schema](
					orderedmap.Pair[string, *schema.Schema]{
						Key: "addPropFalse",
						Value: &schema.Schema{
							Type:                 "object",
							AdditionalProperties: false,
							Properties: orderedmap.New[string, *schema.Schema](orderedmap.WithInitialData[string, *schema.Schema](
								orderedmap.Pair[string, *schema.Schema]{
									Key: "foo",
									Value: &schema.Schema{
										Type: "string",
									},
								},
							)),
						},
					},
					orderedmap.Pair[string, *schema.Schema]{
						Key: "addPropTrue",
						Value: &schema.Schema{
							Type:                 "object",
							AdditionalProperties: true,
						},
					},
					orderedmap.Pair[string, *schema.Schema]{
						Key: "addPropSchema",
						Value: &schema.Schema{
							Type: "object",
							AdditionalProperties: &schema.Schema{
								Type: "object",
								Properties: orderedmap.New[string, *schema.Schema](orderedmap.WithInitialData[string, *schema.Schema](
									orderedmap.Pair[string, *schema.Schema]{
										Key: "bar",
										Value: &schema.Schema{
											Type: "string",
										},
									},
								)),
							},
						},
					},
				)),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			want, err := os.ReadFile(filepath.Join("testdata", tc.name+".json"))
			if err != nil {
				t.Fatalf("%d, unexpected error", err)
			}

			got, err := json.MarshalIndent(tc.schema, "", "    ")
			if err != nil {
				t.Fatalf("%d, unexpected error", err)
			}

			require.JSONEq(t, string(got), string(want))
		})
	}
}

func TestUnmarshal(t *testing.T) {
	type testData struct {
		name string
		want schema.Schema
	}
	tests := []testData{
		{
			name: "schema",
			want: schema.Schema{
				Properties: orderedmap.New[string, *schema.Schema](orderedmap.WithInitialData[string, *schema.Schema](
					orderedmap.Pair[string, *schema.Schema]{
						Key: "addPropFalse",
						Value: &schema.Schema{
							Type:                 "object",
							AdditionalProperties: false,
							Properties: orderedmap.New[string, *schema.Schema](orderedmap.WithInitialData[string, *schema.Schema](
								orderedmap.Pair[string, *schema.Schema]{
									Key: "foo",
									Value: &schema.Schema{
										Type: "string",
									},
								},
							)),
						},
					},
					orderedmap.Pair[string, *schema.Schema]{
						Key: "addPropTrue",
						Value: &schema.Schema{
							Type:                 "object",
							AdditionalProperties: true,
						},
					},
					orderedmap.Pair[string, *schema.Schema]{
						Key: "addPropSchema",
						Value: &schema.Schema{
							Type: "object",
							AdditionalProperties: &schema.Schema{
								Type: "object",
								Properties: orderedmap.New[string, *schema.Schema](orderedmap.WithInitialData[string, *schema.Schema](
									orderedmap.Pair[string, *schema.Schema]{
										Key: "bar",
										Value: &schema.Schema{
											Type: "string",
										},
									},
								)),
							},
						},
					},
				)),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bytes, err := os.ReadFile(filepath.Join("testdata", tc.name+".json"))
			if err != nil {
				t.Fatalf("%d, unexpected error", err)
			}

			got := schema.Schema{}
			err = json.Unmarshal(bytes, &got)
			if err != nil {
				t.Fatalf("%d, unexpected error", err)
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got: %#v want %#v", got, tc.want)
			}
		})
	}
}
