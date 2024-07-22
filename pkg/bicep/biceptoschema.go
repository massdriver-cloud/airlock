package bicep

import (
	"encoding/json"
	"errors"
	"slices"

	bp "github.com/Checkmarx/kics/v2/pkg/parser/bicep"
	"github.com/massdriver-cloud/airlock/pkg/schema"

	orderedmap "github.com/wk8/go-ordered-map/v2"
)

func BicepToSchema(valuesPath string) (string, error) {
	parser := bp.Parser{}

	// the top level node is a document node. We need to go one layer
	// deeper to get the actual yaml content
	params := new(schema.Schema)
	params.Type = "object"
	params.Properties = orderedmap.New[string, *schema.Schema]()
	params.Required = []string{}

	doc, _, err := parser.Parse(valuesPath, nil)
	if err != nil {
		return "", err
	}

	for name, value := range doc[0]["parameters"].(map[string]interface{}) {
		dataMap := value.(map[string]interface{})

		property := new(schema.Schema)

		propType, err := parseType(dataMap["type"].(string))
		if err != nil {
			return "", err
		}

		if def, exists := dataMap["defaultValue"]; exists {
			property.Default = def
		}

		if metadata, exists := dataMap["metadata"]; exists {
			if description, exists := metadata.(map[string]interface{})["description"]; exists {
				property.Description = description.(string)
			}
		}

		property.Title = name
		property.Type = propType

		params.Properties.Set(name, property)
		params.Required = append(params.Required, name)
	}
	slices.Sort(params.Required)

	out, err := json.Marshal(params)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func parseType(typeString string) (string, error) {
	switch typeString {
	case "int":
		return "integer", nil
	case "bool":
		return "boolean", nil
	case "string":
		return "string", nil
	case "array":
		return "array", nil
	case "object":
		return "object", nil
	default:
		return "", errors.New("unknown type: " + typeString)
	}
}
