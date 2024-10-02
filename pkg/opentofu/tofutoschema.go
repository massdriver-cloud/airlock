package opentofu

import (
	"errors"
	"slices"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/massdriver-cloud/airlock/pkg/schema"
	"github.com/massdriver-cloud/terraform-config-inspect/tfconfig"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"github.com/zclconf/go-cty/cty"
)

func TofuToSchema(modulePath string) (*schema.Schema, error) {
	module, err := tfconfig.LoadModule(modulePath)
	if err != nil {
		return nil, err
	}

	sch := new(schema.Schema)
	sch.Properties = orderedmap.New[string, *schema.Schema]()

	for _, variable := range module.Variables {
		variableSchema, err := variableToSchema(variable)
		if err != nil {
			return nil, err
		}
		sch.Properties.Set(variable.Name, variableSchema)
		sch.Required = append(sch.Required, variable.Name)
	}

	slices.Sort(sch.Required)

	return sch, nil
}

func variableToSchema(variable *tfconfig.Variable) (*schema.Schema, error) {
	schema := new(schema.Schema)
	variableType, def, err := variableTypeStringToCtyType(variable.Type)
	if err != nil {
		return nil, err
	}
	err = hydrateSchemaFromNameAndType(variable.Name, variableType, schema, def)
	if err != nil {
		return nil, err
	}

	schema.Description = string(variable.Description)

	if variable.Default != nil {
		schema.Default = variable.Default
	}

	if variable.Default == nil && variable.Type == "bool" {
		schema.Default = false
	}

	// fmt.Printf("Variable type: %v\nVariable name: %v\n", variable.Type, variable.Name)

	return schema, nil
}

func variableTypeStringToCtyType(variableType string) (cty.Type, *typeexpr.Defaults, error) {
	expr, diags := hclsyntax.ParseExpression([]byte(variableType), "", hcl.Pos{Line: 1, Column: 1})
	if len(diags) != 0 {
		return cty.NilType, nil, errors.New(diags.Error())
	}
	ty, def, diags := typeexpr.TypeConstraintWithDefaults(expr)
	if len(diags) != 0 {
		return cty.NilType, nil, errors.New(diags.Error())
	}
	return ty, def, nil
}

func hydrateSchemaFromNameAndType(name string, ty cty.Type, schema *schema.Schema, def *typeexpr.Defaults) error {
	if ty.IsPrimitiveType() {
		hydratePrimitiveSchema(name, ty, schema, def)
	} else if ty.IsMapType() {
		hydrateMapSchema(name, ty, schema, def)
	} else if ty.IsObjectType() {
		hydrateObjectSchema(name, ty, schema, def)
	} else if ty.IsListType() {
		hydrateArraySchema(name, ty, schema, def)
	} else if ty.IsSetType() {
		hydrateSetSchema(name, ty, schema, def)
	}
	return nil
}

func hydratePrimitiveSchema(name string, ty cty.Type, schema *schema.Schema, def *typeexpr.Defaults) {
	schema.Title = name
	switch ty {
	case cty.String:
		schema.Type = "string"
		if def != nil {
			if defVal, exists := def.DefaultValues[name]; exists {
				schema.Default = defVal.AsString()
			}
		}
	case cty.Bool:
		schema.Type = "boolean"
		if def != nil {
			if defVal, exists := def.DefaultValues[name]; exists {
				if defVal.True() {
					schema.Default = true
				} else {
					schema.Default = false
				}
			}
		}
	case cty.Number:
		schema.Type = "number"
		if def != nil {
			if defVal, exists := def.DefaultValues[name]; exists {
				defNum, _ := defVal.AsBigFloat().Float64()
				schema.Default = defNum
			}
		}
	}
}

func hydrateObjectSchema(name string, ty cty.Type, sch *schema.Schema, def *typeexpr.Defaults) {
	sch.Title = name
	sch.Type = "object"
	sch.Properties = orderedmap.New[string, *schema.Schema]()
	for attName, attType := range ty.AttributeTypes() {
		var nestDef *typeexpr.Defaults
		if def != nil {
			if defVal, exists := def.Children[attName]; exists {
				nestDef = defVal
			} else {
				nestDef = nil
			}
		}
		attributeSchema := new(schema.Schema)
		hydrateSchemaFromNameAndType(attName, attType, attributeSchema, nestDef)
		sch.Properties.Set(attName, attributeSchema)
		if !ty.AttributeOptional(attName) {
			sch.Required = append(sch.Required, attName)
		}
	}
	slices.Sort(sch.Required)
}

func hydrateMapSchema(name string, ty cty.Type, sch *schema.Schema, def *typeexpr.Defaults) {
	sch.Title = name
	sch.Type = "object"
	sch.PropertyNames = &schema.Schema{
		Pattern: "^.*$",
	}
	sch.AdditionalProperties = new(schema.Schema)
	hydrateSchemaFromNameAndType("", ty.ElementType(), sch.AdditionalProperties.(*schema.Schema), def)
}

func hydrateArraySchema(name string, ty cty.Type, sch *schema.Schema, def *typeexpr.Defaults) {
	sch.Title = name
	sch.Type = "array"
	sch.Items = new(schema.Schema)
	hydrateSchemaFromNameAndType("", ty.ElementType(), sch.Items, def)
}

func hydrateSetSchema(name string, ty cty.Type, sch *schema.Schema, def *typeexpr.Defaults) {
	hydrateArraySchema(name, ty, sch, def)
	sch.UniqueItems = true
	hydrateSchemaFromNameAndType("", ty.ElementType(), sch.Items, def)
}
