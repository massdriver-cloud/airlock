package opentofu

import (
	"encoding/json"
	"io"
	"reflect"
	"slices"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/massdriver-cloud/airlock/pkg/schema"
)

func SchemaToTofu(in io.Reader) ([]byte, error) {
	bytes, err := io.ReadAll(in)
	if err != nil {
		return nil, err
	}

	root := schema.Schema{}
	err = json.Unmarshal(bytes, &root)
	if err != nil {
		return nil, err
	}

	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	flattenedProperties := schema.ExpandProperties(&root)
	for prop := flattenedProperties.Oldest(); prop != nil; prop = prop.Next() {
		required := slices.Contains(root.Required, prop.Key)
		err = createTopLevelVariableBlock(prop.Key, prop.Value, required, rootBody)
		if err != nil {
			return nil, err
		}
	}

	return f.Bytes(), nil
}

func createTopLevelVariableBlock(name string, param *schema.Schema, required bool, body *hclwrite.Body) error {
	varBlock := body.AppendNewBlock("variable", []string{name})
	varBody := varBlock.Body()

	varBody.SetAttributeRaw(
		"type",
		typeExprTokens(param, false),
	)

	// If this value isn't required, then we should set the default
	if !required {
		defaultValue, err := json.Marshal(param.Default)
		if err != nil {
			return err
		}

		varBody.SetAttributeRaw(
			"default",
			hclwrite.TokensForIdentifier(string(defaultValue)),
		)
	}
	return nil
}

func typeExprTokens(node *schema.Schema, optional bool) hclwrite.Tokens {
	if optional {
		return hclwrite.TokensForFunctionCall("optional", typeExprTokens(node, false))
	}

	switch node.Type {
	case "string":
		return hclwrite.TokensForIdentifier("string")
	case "boolean":
		return hclwrite.TokensForIdentifier("bool")
	case "integer", "number":
		return hclwrite.TokensForIdentifier("number")
	case "array":
		return convertArray(node)
	case "object":
		return convertObject(node)
	default:
		return hclwrite.TokensForIdentifier("any")
	}
}

func convertArray(node *schema.Schema) hclwrite.Tokens {
	return hclwrite.TokensForFunctionCall("list", typeExprTokens(node.Items, false))
}

func convertMap(node *schema.Schema) hclwrite.Tokens {
	// opentofu maps must all have the same type for the map value. Therefore there are only limited
	// cases where we can interpret a map. Otherwise, we have to give up and just use type "any"

	// first check: if there are any existing properties, we can't guarantee everything is the same type, so bail
	if node.Properties.Len() > 0 {
		return hclwrite.TokensForIdentifier("any")
	}

	// AdditionalProperties is a strange field that can be either a boolean or a schema. We treat nil and false the same
	additionalPropertiesIsFalseOrNull := node.AdditionalProperties == nil || (reflect.TypeOf(node.AdditionalProperties).Kind() == reflect.Bool && !node.AdditionalProperties.(bool))

	// if additionalProperties is a schema (and no other dynamic is set), we can interpret as a map
	if !additionalPropertiesIsFalseOrNull && node.PatternProperties == nil {
		// if its just a boolean "true", then type is "any"
		if reflect.TypeOf(node.AdditionalProperties).Kind() == reflect.Bool && node.AdditionalProperties.(bool) {
			return hclwrite.TokensForIdentifier("any")
		}

		// its a schema, so we can parse it and convert it as a map
		addPropSchema, ok := node.AdditionalProperties.(*schema.Schema)
		if ok {
			return hclwrite.TokensForFunctionCall("map", typeExprTokens(addPropSchema, false))
		}
	}

	// if there is exactly 1 patternProperty specification (and no other dynamic is set), we can interpret as a map
	if node.PatternProperties != nil && len(node.PatternProperties) == 1 && additionalPropertiesIsFalseOrNull {
		for _, patternProp := range node.PatternProperties {
			return hclwrite.TokensForFunctionCall("map", typeExprTokens(patternProp, false))
		}
	}

	return hclwrite.TokensForIdentifier("any")
}

func convertObject(node *schema.Schema) hclwrite.Tokens {
	// if any of the fields that imply there are dynamic properties exists, we need to try to interpret as a map
	additionalPropertiesIsFalseOrNull := node.AdditionalProperties == nil || (reflect.TypeOf(node.AdditionalProperties).Kind() == reflect.Bool && !node.AdditionalProperties.(bool))
	if !additionalPropertiesIsFalseOrNull || node.PatternProperties != nil {
		return convertMap(node)
	}

	items := parseObject(node)

	return hclwrite.TokensForFunctionCall("object", hclwrite.TokensForObject(items))
}

func parseObject(node *schema.Schema) []hclwrite.ObjectAttrTokens {
	items := []hclwrite.ObjectAttrTokens{}

	flattenedProperties := schema.ExpandProperties(node)
	for prop := flattenedProperties.Oldest(); prop != nil; prop = prop.Next() {
		optional := !slices.Contains(node.Required, prop.Key)
		items = append(items, hclwrite.ObjectAttrTokens{
			Name:  hclwrite.TokensForIdentifier(prop.Key),
			Value: typeExprTokens(prop.Value, optional),
		})
	}

	return items
}
