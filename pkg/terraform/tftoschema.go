package terraform

import (
	"errors"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/massdriver-cloud/airlock/pkg/schema"
	"github.com/terraform-docs/terraform-docs/print"
	tfd "github.com/terraform-docs/terraform-docs/terraform"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"github.com/zclconf/go-cty/cty"
)

func TfToSchema(modulePath string) (*schema.Schema, error) {
	config := print.NewConfig()
	config.ModuleRoot = modulePath

	module, err := tfd.LoadWithOptions(config)
	if err != nil {
		return nil, err
	}

	sch := new(schema.Schema)
	sch.Properties = orderedmap.New[string, *schema.Schema]()

	for _, variable := range module.Inputs {
		variableSchema, err := variableToSchema(variable)
		if err != nil {
			return nil, err
		}
		sch.Properties.Set(variable.Name, variableSchema)
		sch.Required = append(sch.Required, variable.Name)
	}

	return sch, nil
}

func variableToSchema(variable *tfd.Input) (*schema.Schema, error) {
	schema := new(schema.Schema)
	variableType, err := variableTypeStringToCtyType(string(variable.Type))
	if err != nil {
		return nil, err
	}
	err = hydrateSchemaFromNameAndType(variable.Name, variableType, schema)
	if err != nil {
		return nil, err
	}

	schema.Description = string(variable.Description)

	if variable.Default.Raw() != nil {
		schema.Default = variable.Default
	}

	return schema, nil
}

func variableTypeStringToCtyType(variableType string) (cty.Type, error) {
	expr, diags := hclsyntax.ParseExpression([]byte(variableType), "", hcl.Pos{Line: 1, Column: 1})
	if len(diags) != 0 {
		return cty.NilType, errors.New(diags.Error())
	}
	ty, diags := typeexpr.TypeConstraint(expr)
	if len(diags) != 0 {
		return cty.NilType, errors.New(diags.Error())
	}
	return ty, nil
}

func hydrateSchemaFromNameAndType(name string, ty cty.Type, schema *schema.Schema) error {
	if ty.IsPrimitiveType() {
		hydratePrimitiveSchema(name, ty, schema)
	} else if ty.IsMapType() {
		hydrateMapSchema(name, ty, schema)
	} else if ty.IsObjectType() {
		hydrateObjectSchema(name, ty, schema)
	} else if ty.IsListType() {
		hydrateArraySchema(name, ty, schema)
	} else if ty.IsSetType() {
		hydrateSetSchema(name, ty, schema)
	}
	return nil
}

func hydratePrimitiveSchema(name string, ty cty.Type, schema *schema.Schema) {
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

func hydrateObjectSchema(name string, ty cty.Type, sch *schema.Schema) {
	sch.Title = name
	sch.Type = "object"
	sch.Properties = orderedmap.New[string, *schema.Schema]()
	for attName, attType := range ty.AttributeTypes() {
		attributeSchema := new(schema.Schema)
		hydrateSchemaFromNameAndType(attName, attType, attributeSchema)
		sch.Properties.Set(attName, attributeSchema)
		if !ty.AttributeOptional(attName) {
			sch.Required = append(sch.Required, attName)
		}
	}
}

func hydrateMapSchema(name string, ty cty.Type, sch *schema.Schema) {
	sch.Title = name
	sch.Type = "object"
	sch.PropertyNames = &schema.Schema{
		Pattern: "^.*$",
	}
	sch.AdditionalProperties = new(schema.Schema)
	hydrateSchemaFromNameAndType("", ty.ElementType(), sch.AdditionalProperties.(*schema.Schema))
}

func hydrateArraySchema(name string, ty cty.Type, sch *schema.Schema) {
	sch.Title = name
	sch.Type = "array"
	sch.Items = new(schema.Schema)
	hydrateSchemaFromNameAndType("", ty.ElementType(), sch.Items)
}

func hydrateSetSchema(name string, ty cty.Type, sch *schema.Schema) {
	hydrateArraySchema(name, ty, sch)
	sch.UniqueItems = true
	hydrateSchemaFromNameAndType("", ty.ElementType(), sch.Items)
}
