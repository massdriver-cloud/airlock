package terraform

import (
	"encoding/json"
	"errors"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/iancoleman/orderedmap"
	"github.com/invopop/jsonschema"
	"github.com/terraform-docs/terraform-docs/print"
	tfd "github.com/terraform-docs/terraform-docs/terraform"
	"github.com/zclconf/go-cty/cty"
)

func Run(modulePath string) (string, error) {
	config := print.NewConfig()
	config.ModuleRoot = modulePath

	module, err := tfd.LoadWithOptions(config)
	if err != nil {
		return "", err
	}

	params := new(jsonschema.Schema)
	params.Properties = orderedmap.New()

	for _, variable := range module.Inputs {
		variableSchema, err := variableToSchema(variable)
		if err != nil {
			return "", err
		}
		params.Properties.Set(variable.Name, variableSchema)
		params.Required = append(params.Required, variable.Name)
	}

	bytes, err := json.MarshalIndent(params, "", "  ")
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func variableToSchema(variable *tfd.Input) (*jsonschema.Schema, error) {
	schema := new(jsonschema.Schema)
	variableType, err := variableTypeStringToCtyType(string(variable.Type))
	if err != nil {
		return nil, err
	}
	hydrateSchemaFromNameAndType(variable.Name, variableType, schema)

	schema.Description = string(variable.Description)
	schema.Default = variable.Default

	return schema, nil
}

func variableTypeStringToCtyType(variableType string) (cty.Type, error) {
	expr, diags := hclsyntax.ParseExpression([]byte(variableType), "whatever.foo", hcl.Pos{Line: 1, Column: 1})
	if len(diags) != 0 {
		return cty.NilType, errors.New(diags.Error())
	}
	ty, _, diags := typeexpr.TypeConstraintWithDefaults(expr)
	if len(diags) != 0 {
		return cty.NilType, errors.New(diags.Error())
	}
	return ty, nil
}

func hydrateSchemaFromNameAndType(name string, ty cty.Type, schema *jsonschema.Schema) error {
	if ty.IsPrimitiveType() {
		hydratePrimitiveSchema(name, ty, schema)
	} else if ty.IsMapType() {
		hydrateMapSchema(name, ty, schema)
	} else if ty.IsObjectType() {
		hydrateObjectSchema(name, ty, schema)
	} else if ty.IsCollectionType() {
		hydrateArraySchema(name, ty, schema)
	}
	return nil
}

func hydratePrimitiveSchema(name string, ty cty.Type, schema *jsonschema.Schema) {
	schema.Title = name
	switch ty {
	case cty.String:
		schema.Type = "string"
	case cty.Bool:
		schema.Type = "boolean"
	case cty.Number:
		schema.Type = "number"
	}
}

func hydrateObjectSchema(name string, ty cty.Type, schema *jsonschema.Schema) {
	schema.Title = name
	schema.Type = "object"
	schema.Properties = orderedmap.New()
	for attName, attType := range ty.AttributeTypes() {
		attributeSchema := new(jsonschema.Schema)
		hydrateSchemaFromNameAndType(attName, attType, attributeSchema)
		schema.Properties.Set(attName, attributeSchema)
		if !ty.AttributeOptional(attName) {
			schema.Required = append(schema.Required, attName)
		}
	}
}

func hydrateMapSchema(name string, ty cty.Type, schema *jsonschema.Schema) {
	schema.Title = name
	schema.Type = "object"
	schema.AdditionalProperties = new(jsonschema.Schema)
	hydrateSchemaFromNameAndType("", ty.ElementType(), schema.AdditionalProperties)
}

func hydrateArraySchema(name string, ty cty.Type, schema *jsonschema.Schema) {
	schema.Title = name
	schema.Type = "array"
	schema.Items = new(jsonschema.Schema)
	hydrateSchemaFromNameAndType("", ty.ElementType(), schema.Items)
}
