package terraform

import (
	"encoding/json"
	"io"
	"reflect"
	"slices"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/massdriver-cloud/airlock/pkg/schema"
)

func SchemaToTf(in io.Reader) ([]byte, error) {
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

	for prop := root.Properties.Oldest(); prop != nil; prop = prop.Next() {
		varBlock := rootBody.AppendNewBlock("variable", []string{prop.Key})
		varBody := varBlock.Body()

		varBody.SetAttributeRaw(
			"type",
			typeExprTokens(prop.Value, false),
		)
	}

	return f.Bytes(), nil
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
	// terraform maps must all have the same type for the map value. Therefore there are only limited
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

	// standard properties
	props := extractObjectProperties(node, false)
	items = mergeProperties(items, props)

	// oneOf, anyOf, allOf
	for _, item := range node.OneOf {
		oneOf := extractObjectProperties(item, true)
		items = mergeProperties(items, oneOf)
	}
	for _, item := range node.AnyOf {
		anyOf := extractObjectProperties(item, true)
		items = mergeProperties(items, anyOf)
	}
	for _, item := range node.AllOf {
		allOf := extractObjectProperties(item, true)
		items = mergeProperties(items, allOf)
	}

	// dependencies
	for _, item := range node.Dependencies {
		dep := parseObject(item)
		items = mergeProperties(items, dep)
	}

	return items
}

func extractObjectProperties(node *schema.Schema, allOptional bool) []hclwrite.ObjectAttrTokens {
	properties := []hclwrite.ObjectAttrTokens{}
	var optional bool
	for prop := node.Properties.Oldest(); prop != nil; prop = prop.Next() {
		if allOptional {
			optional = true
		} else {
			optional = !slices.Contains(node.Required, prop.Key)
		}
		properties = append(properties, hclwrite.ObjectAttrTokens{
			Name:  hclwrite.TokensForIdentifier(prop.Key),
			Value: typeExprTokens(prop.Value, optional),
		})
	}
	return properties
}

func mergeProperties(existing, new []hclwrite.ObjectAttrTokens) []hclwrite.ObjectAttrTokens {
	for _, att := range new {
		exists := false
		for _, exist := range existing {
			if string(exist.Name.Bytes()) == string(att.Name.Bytes()) {
				exists = true
				break
			}
		}
		if !exists {
			existing = append(existing, att)
		}
	}
	return existing
}
