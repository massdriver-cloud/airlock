package bicep

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"

	bp "github.com/Checkmarx/kics/v2/pkg/parser/bicep"
	"github.com/massdriver-cloud/airlock/pkg/schema"

	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type bicepParam struct {
	TypeString    string             `json:"type"`
	DefaultValue  interface{}        `json:"defaultValue"`
	AllowedValues []interface{}      `json:"allowedValues"`
	Metadata      bicepParamMetadata `json:"metadata"`
	MinLength     *uint64            `json:"minLength"`
	MinValue      *uint64            `json:"minValue"`
	MaxLength     *uint64            `json:"maxLength"`
	MaxValue      *uint64            `json:"maxValue"`
}

type bicepParamMetadata struct {
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
}

func BicepToSchema(valuesPath string) (string, error) {
	// using the github.com/Checkmarx/kics parser since he already did the heavy lifting to parse a bicep template
	parser := bp.Parser{}

	params := new(schema.Schema)
	params.Type = "object"
	params.Properties = orderedmap.New[string, *schema.Schema]()
	params.Required = []string{}

	doc, _, err := parser.Parse(valuesPath, nil)
	if err != nil {
		return "", err
	}

	for name, value := range doc[0]["parameters"].(map[string]interface{}) {
		bicepParam := bicepParam{}

		// marshall to json and unmarshall into custom struct to make bicep param easier to access
		bytes, err := json.Marshal(value)
		if err != nil {
			return "", err
		}
		err = json.Unmarshal(bytes, &bicepParam)
		if err != nil {
			return "", err
		}

		property := new(schema.Schema)
		property.Title = name
		property.Description = bicepParam.Metadata.Description

		err = parseBicepParam(property, bicepParam)
		if err != nil {
			return "", err
		}

		params.Properties.Set(name, property)
		params.Required = append(params.Required, name)
	}
	// sorting this here just to help with testing. The order doesn't matter, but to our test suite it does.
	slices.Sort(params.Required)

	out, err := json.Marshal(params)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func parseBicepParam(sch *schema.Schema, bicepParam bicepParam) error {
	switch bicepParam.TypeString {
	case "int":
		return parseIntParam(sch, bicepParam)
	case "bool":
		return parseBoolParam(sch, bicepParam)
	case "string":
		return parseStringParam(sch, bicepParam, false)
	case "secureString":
		return parseStringParam(sch, bicepParam, true)
	case "array":
		return parseArrayParam(sch, bicepParam)
	case "object", "secureObject":
		return parseObjectParam(sch, bicepParam)
	default:
		return errors.New("unknown type: " + bicepParam.TypeString)
	}
}

func parseIntParam(sch *schema.Schema, bicepParam bicepParam) error {
	sch.Type = "integer"
	sch.Default = bicepParam.DefaultValue

	if bicepParam.AllowedValues != nil && len(bicepParam.AllowedValues) == 1 {
		sch.Enum = bicepParam.AllowedValues[0].([]interface{})
	}

	if bicepParam.MinValue != nil {
		sch.Minimum = json.Number(fmt.Sprintf("%d", *bicepParam.MinValue))
	}
	if bicepParam.MaxValue != nil {
		sch.Maximum = json.Number(fmt.Sprintf("%d", *bicepParam.MaxValue))
	}

	return nil
}

func parseBoolParam(sch *schema.Schema, bicepParam bicepParam) error {
	sch.Type = "boolean"
	sch.Default = bicepParam.DefaultValue
	return nil
}

func parseStringParam(sch *schema.Schema, bicepParam bicepParam, secure bool) error {
	sch.Type = "string"
	sch.Default = bicepParam.DefaultValue

	if secure {
		sch.Format = "password"
	}

	if bicepParam.AllowedValues != nil && len(bicepParam.AllowedValues) == 1 {
		sch.Enum = bicepParam.AllowedValues[0].([]interface{})
	}

	sch.MinLength = bicepParam.MinLength
	sch.MaxLength = bicepParam.MaxLength

	return nil
}

func parseArrayParam(sch *schema.Schema, bicepParam bicepParam) error {
	sch.Type = "array"

	sch.MinItems = bicepParam.MinLength
	sch.MaxItems = bicepParam.MaxLength

	if bicepParam.DefaultValue != nil && len(bicepParam.DefaultValue.([]interface{})) != 0 {
		parseArrayType(sch, bicepParam.DefaultValue.([]interface{}))
	}
	return nil
}

func parseObjectParam(sch *schema.Schema, bicepParam bicepParam) error {
	sch.Type = "object"

	if bicepParam.DefaultValue != nil && len(bicepParam.DefaultValue.(map[string]interface{})) > 1 {
		parseObjectType(sch, bicepParam.DefaultValue.(map[string]interface{}))
	}
	return nil
}

func parseObjectType(sch *schema.Schema, objValue map[string]interface{}) error {
	sch.Properties = orderedmap.New[string, *schema.Schema]()
	sch.Required = []string{}

	for name, value := range objValue {
		if name == "_kics_lines" {
			continue
		}

		property := new(schema.Schema)
		property.Title = name

		switch reflect.TypeOf(value).Kind() {
		case reflect.Float64:
			property.Type = "integer"
			property.Default = value
		case reflect.Bool:
			property.Type = "boolean"
			property.Default = value
		case reflect.String:
			property.Type = "string"
			property.Default = value
		case reflect.Slice:
			property.Type = "array"
			parseArrayType(property, value.([]interface{}))
		case reflect.Map:
			property.Type = "object"
			parseObjectType(property, value.(map[string]interface{}))
		default:
			return errors.New("unknown type: " + reflect.TypeOf(value).String())
		}

		sch.Properties.Set(name, property)
		sch.Required = append(sch.Required, name)
		slices.Sort(sch.Required)
	}

	return nil
}

func parseArrayType(sch *schema.Schema, value []interface{}) error {
	items := new(schema.Schema)

	elem := value[0]
	switch reflect.TypeOf(elem).Kind() {
	case reflect.Float64:
		items.Type = "integer"
		sch.Default = value
	case reflect.Bool:
		items.Type = "boolean"
		sch.Default = value
	case reflect.String:
		items.Type = "string"
		sch.Default = value
	case reflect.Slice:
		items.Type = "array"
		parseArrayType(items, elem.([]interface{}))
	case reflect.Map:
		items.Type = "object"
		parseObjectType(items, elem.(map[string]interface{}))
	default:
		return errors.New("unknown type: " + reflect.TypeOf(elem).String())
	}

	sch.Items = items

	return nil
}
