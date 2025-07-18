package opentofu

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/massdriver-cloud/airlock/pkg/schema"
	"github.com/massdriver-cloud/terraform-config-inspect/tfconfig"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
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
			return nil, fmt.Errorf("failed to convert variable %q to schema: %w", variable.Name, err)
		}
		sch.Properties.Set(variable.Name, variableSchema)
		sch.Required = append(sch.Required, variable.Name)
	}

	slices.Sort(sch.Required)

	return sch, nil
}

func variableToSchema(variable *tfconfig.Variable) (*schema.Schema, error) {
	schema := new(schema.Schema)
	variableType, defaults, typeErr := variableTypeStringToCtyType(variable.Type)
	if typeErr != nil {
		return nil, fmt.Errorf("failed to parse type %q: %w", variable.Type, typeErr)
	}
	// To simplify the logic of recursively walking the Defaults structure in objects types,
	// we make the extracted Defaults a Child of a dummy "top level" node
	var topLevelDefault *typeexpr.Defaults
	if defaults != nil {
		topLevelDefault = new(typeexpr.Defaults)
		topLevelDefault.Children = map[string]*typeexpr.Defaults{
			variable.Name: defaults,
		}
	}
	if hydrateErr := hydrateSchemaFromNameTypeAndDefaults(schema, variable.Name, variableType, topLevelDefault); hydrateErr != nil {
		return nil, fmt.Errorf("failed to hydrate schema for variable %q: %w", variable.Name, hydrateErr)
	}

	schema.Description = variable.Description

	if variable.Default != nil {
		schema.Default = variable.Default
	}

	if variable.Default == nil && variable.Type == "bool" {
		schema.Default = false
	}

	return schema, nil
}

func variableTypeStringToCtyType(variableType string) (cty.Type, *typeexpr.Defaults, error) {
	if variableType == "" {
		return cty.NilType, nil, errors.New("type cannot be empty")
	}
	if variableType == "any" {
		return cty.NilType, nil, errors.New("type 'any' cannot be converted to a JSON schema type")
	}
	expr, diags := hclsyntax.ParseExpression([]byte(variableType), "", hcl.Pos{Line: 1, Column: 1})
	if len(diags) != 0 {
		return cty.NilType, nil, errors.New(diags.Error())
	}
	ty, defaults, diags := typeexpr.TypeConstraintWithDefaults(expr)
	if len(diags) != 0 {
		return cty.NilType, nil, errors.New(diags.Error())
	}
	return ty, defaults, nil
}

func hydrateSchemaFromNameTypeAndDefaults(sch *schema.Schema, name string, ty cty.Type, defaults *typeexpr.Defaults) error {
	sch.Title = name

	if defaults != nil {
		if defVal, exists := defaults.DefaultValues[name]; exists {
			sch.Default = ctyValueToInterface(defVal)
		}
	}

	//nolint:gocritic
	if ty.IsPrimitiveType() {
		hydratePrimitiveSchema(sch, ty)
	} else if ty.IsMapType() {
		return hydrateMapSchema(sch, name, ty, defaults)
	} else if ty.IsObjectType() {
		return hydrateObjectSchema(sch, name, ty, defaults)
	} else if ty.IsListType() {
		return hydrateArraySchema(sch, name, ty, defaults)
	} else if ty.IsSetType() {
		return hydrateSetSchema(sch, name, ty, defaults)
	} else if ty.HasDynamicTypes() {
		return fmt.Errorf("dynamic types are not supported (are you using type 'any'?)")
	} else {
		return fmt.Errorf("unsupported type %q", ty.FriendlyName())
	}
	return nil
}

func hydratePrimitiveSchema(sch *schema.Schema, ty cty.Type) {
	switch ty {
	case cty.String:
		sch.Type = "string"
	case cty.Bool:
		sch.Type = "boolean"
	case cty.Number:
		sch.Type = "number"
	}
}

func hydrateObjectSchema(sch *schema.Schema, name string, ty cty.Type, defaults *typeexpr.Defaults) error {
	sch.Type = "object"
	sch.Properties = orderedmap.New[string, *schema.Schema]()
	for attName, attType := range ty.AttributeTypes() {
		attributeSchema := new(schema.Schema)
		if err := hydrateSchemaFromNameTypeAndDefaults(attributeSchema, attName, attType, getDefaultChildren(name, defaults)); err != nil {
			return err
		}
		sch.Properties.Set(attName, attributeSchema)
		if !ty.AttributeOptional(attName) {
			sch.Required = append(sch.Required, attName)
		}
	}
	slices.Sort(sch.Required)
	return nil
}

func hydrateMapSchema(sch *schema.Schema, name string, ty cty.Type, defaults *typeexpr.Defaults) error {
	sch.Type = "object"
	sch.PropertyNames = &schema.Schema{
		Pattern: "^.*$",
	}
	sch.AdditionalProperties = new(schema.Schema)
	return hydrateSchemaFromNameTypeAndDefaults(sch.AdditionalProperties.(*schema.Schema), "", ty.ElementType(), getDefaultChildren(name, defaults))
}

func hydrateArraySchema(sch *schema.Schema, name string, ty cty.Type, defaults *typeexpr.Defaults) error {
	sch.Type = "array"
	sch.Items = new(schema.Schema)
	return hydrateSchemaFromNameTypeAndDefaults(sch.Items, "", ty.ElementType(), getDefaultChildren(name, defaults))
}

func hydrateSetSchema(sch *schema.Schema, name string, ty cty.Type, defaults *typeexpr.Defaults) error {
	sch.UniqueItems = true
	return hydrateArraySchema(sch, name, ty, defaults)
}

func ctyValueToInterface(val cty.Value) interface{} {
	valJSON, err := ctyjson.Marshal(val, val.Type())
	if err != nil {
		// Should never happen, since all possible known
		// values have a JSON mapping.
		panic(fmt.Errorf("failed to serialize default value as JSON: %w", err))
	}
	var def interface{}
	err = json.Unmarshal(valJSON, &def)
	if err != nil {
		// Again should never happen, because valJSON is
		// guaranteed valid by ctyjson.Marshal.
		panic(fmt.Errorf("failed to re-parse default value from JSON: %w", err))
	}
	removeNullKeys(def)
	return def
}

func getDefaultChildren(name string, defaults *typeexpr.Defaults) *typeexpr.Defaults {
	var children *typeexpr.Defaults
	if defaults != nil {
		if attDefaultVal, exists := defaults.Children[name]; exists {
			children = attDefaultVal
		}
	}
	return children
}

// if fields are missing from the default value for an object in the HCL, they are set to null
// we want to remove these fields from the default instead of creating a null default in the schema
func removeNullKeys(defVal interface{}) {
	assertedDefVal, ok := defVal.(map[string]interface{})
	if !ok {
		return
	}
	for key, value := range assertedDefVal {
		if value == nil {
			delete(assertedDefVal, key)
			continue
		}
		if valObj, objOk := assertedDefVal[key].(map[string]interface{}); objOk {
			removeNullKeys(valObj)
		}
	}
}
