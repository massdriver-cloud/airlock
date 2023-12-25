package schema_test

import (
	"reflect"
	"testing"

	"github.com/massdriver-cloud/airlock/pkg/schema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

func TestExpandProperties(t *testing.T) {
	type testData struct {
		name   string
		schema *schema.Schema
		want   *orderedmap.OrderedMap[string, *schema.Schema]
	}
	tests := []testData{
		{
			name: "simple",
			schema: &schema.Schema{
				Properties: orderedmap.New[string, *schema.Schema](orderedmap.WithInitialData[string, *schema.Schema](
					orderedmap.Pair[string, *schema.Schema]{
						Key: "foo",
						Value: &schema.Schema{
							Type: "string",
						},
					},
				)),
			},
			want: orderedmap.New[string, *schema.Schema](orderedmap.WithInitialData[string, *schema.Schema](
				orderedmap.Pair[string, *schema.Schema]{
					Key: "foo",
					Value: &schema.Schema{
						Type: "string",
					},
				},
			)),
		},
		{
			name: "oneOf",
			schema: &schema.Schema{
				OneOf: []*schema.Schema{
					{
						Properties: orderedmap.New[string, *schema.Schema](orderedmap.WithInitialData[string, *schema.Schema](
							orderedmap.Pair[string, *schema.Schema]{
								Key: "foo",
								Value: &schema.Schema{
									Type: "string",
								},
							},
						)),
					},
					{
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
			want: orderedmap.New[string, *schema.Schema](orderedmap.WithInitialData[string, *schema.Schema](
				orderedmap.Pair[string, *schema.Schema]{
					Key: "foo",
					Value: &schema.Schema{
						Type: "string",
					},
				},
				orderedmap.Pair[string, *schema.Schema]{
					Key: "bar",
					Value: &schema.Schema{
						Type: "string",
					},
				},
			)),
		},
		{
			name: "dependency",
			schema: &schema.Schema{
				Properties: orderedmap.New[string, *schema.Schema](orderedmap.WithInitialData[string, *schema.Schema](
					orderedmap.Pair[string, *schema.Schema]{
						Key: "foo",
						Value: &schema.Schema{
							Type: "string",
						},
					},
				)),
				Dependencies: map[string]*schema.Schema{
					"foo": {
						OneOf: []*schema.Schema{
							{
								Properties: orderedmap.New[string, *schema.Schema](orderedmap.WithInitialData[string, *schema.Schema](
									orderedmap.Pair[string, *schema.Schema]{
										Key: "foo",
										Value: &schema.Schema{
											Const: "blah",
										},
									},
								)),
							},
							{
								Properties: orderedmap.New[string, *schema.Schema](orderedmap.WithInitialData[string, *schema.Schema](
									orderedmap.Pair[string, *schema.Schema]{
										Key: "foo",
										Value: &schema.Schema{
											Const: "blargh",
										},
									},
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
				},
			},
			want: orderedmap.New[string, *schema.Schema](orderedmap.WithInitialData[string, *schema.Schema](
				orderedmap.Pair[string, *schema.Schema]{
					Key: "foo",
					Value: &schema.Schema{
						Type: "string",
					},
				},
				orderedmap.Pair[string, *schema.Schema]{
					Key: "bar",
					Value: &schema.Schema{
						Type: "string",
					},
				},
			)),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// want, err := os.ReadFile(filepath.Join("testdata/schemas", tc.name+".tf"))
			// if err != nil {
			// 	t.Fatalf("%d, unexpected error", err)
			// }

			// schemaFile, err := os.Open(filepath.Join("testdata/schemas", tc.name+".json"))
			// if err != nil {
			// 	t.Fatalf("%d, unexpected error", err)
			// }

			got := schema.ExpandProperties(tc.schema)
			// if err != nil {
			// 	t.Fatalf("%d, unexpected error", err)
			// }

			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %#v want %#v", got, tc.want)
			}
		})
	}
}
