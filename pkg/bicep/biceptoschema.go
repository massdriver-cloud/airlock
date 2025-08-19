package bicep

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"

	bp "github.com/Checkmarx/kics/v2/pkg/parser/bicep"
	"github.com/massdriver-cloud/airlock/pkg/result"
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

func BicepToSchema(templatePath string) result.SchemaResult {
	// using the github.com/Checkmarx/kics parser since he already did the heavy lifting to parse a bicep template
	parser := bp.Parser{}

	sch := new(schema.Schema)
	sch.Type = "object"
	sch.Properties = orderedmap.New[string, *schema.Schema]()
	sch.Required = []string{}

	doc, _, parseErr := parser.Parse(templatePath, nil)
	if parseErr != nil {
		return result.SchemaResult{
			Schema: nil,
			Diags: []result.Diagnostic{
				{
					Path:    templatePath,
					Code:    "file_read_error",
					Message: fmt.Sprintf("failed to read bicep file: %s", parseErr),
					Level:   result.Error,
				},
			},
		}
	}

	output := result.SchemaResult{
		Schema: sch,
		Diags:  []result.Diagnostic{},
	}

	for name, value := range doc[0]["parameters"].(map[string]interface{}) {
		param := new(bicepParam)

		// marshal to json and unmarshal into custom struct to make bicep param easier to access
		bytes, marshalErr := json.Marshal(value)
		if marshalErr != nil {
			output.Diags = append(output.Diags, result.Diagnostic{
				Path:    name,
				Code:    "invalid_value",
				Message: fmt.Sprintf("failed to marshal bicep param %s: %s", name, marshalErr),
				Level:   result.Error,
			})
			continue
		}
		unmarshalErr := json.Unmarshal(bytes, &param)
		if unmarshalErr != nil {
			output.Diags = append(output.Diags, result.Diagnostic{
				Path:    name,
				Code:    "invalid_value",
				Message: fmt.Sprintf("failed to unmarshal bicep param %s: %s", name, unmarshalErr),
				Level:   result.Error,
			})
			continue
		}

		property := new(schema.Schema)
		property.Title = name
		property.Description = param.Metadata.Description

		output.Diags = parseBicepParam(property, param, output.Diags)

		sch.Properties.Set(name, property)
		sch.Required = append(sch.Required, name)
	}
	// sorting this here just to help with testing. The order doesn't matter, but to our test suite it does.
	slices.Sort(sch.Required)

	return output
}

func parseBicepParam(sch *schema.Schema, bicepParam *bicepParam, diags []result.Diagnostic) []result.Diagnostic {
	switch bicepParam.TypeString {
	case "int":
		return parseIntParam(sch, bicepParam, diags)
	case "bool":
		parseBoolParam(sch, bicepParam)
	case "string":
		return parseStringParam(sch, bicepParam, false, diags)
	case "secureString":
		return parseStringParam(sch, bicepParam, true, diags)
	case "array":
		return parseArrayParam(sch, bicepParam, diags)
	case "object", "secureObject":
		return parseObjectParam(sch, bicepParam, diags)
	default:
		sch.Comment = fmt.Sprintf("Airlock Warning: unknown type from Bicep parameter (%s)", bicepParam.TypeString)
		return append(diags, result.Diagnostic{
			Path:    sch.Title,
			Code:    "unknown_type",
			Message: fmt.Sprintf("type of field %s is unsupported (%s)", sch.Title, bicepParam.TypeString),
			Level:   result.Warning,
		})
	}
	return diags
}

func parseIntParam(sch *schema.Schema, bicepParam *bicepParam, diags []result.Diagnostic) []result.Diagnostic {
	sch.Type = "integer"
	sch.Default = bicepParam.DefaultValue

	allowedVals := bicepParam.AllowedValues
	if len(allowedVals) == 1 {
		assertedEnum, ok := allowedVals[0].([]interface{})
		if ok {
			sch.Enum = assertedEnum
		} else {
			sch.Comment = "Airlock Warning: unable to convert 'allowedValues' to enum"
			diags = append(diags, result.Diagnostic{
				Path:    sch.Title,
				Code:    "invalid_value",
				Message: fmt.Sprintf("unable to convert 'allowedValues' to enum in bicep param %s", sch.Title),
				Level:   result.Warning,
			})
		}

	}

	if bicepParam.MinValue != nil {
		sch.Minimum = json.Number(fmt.Sprintf("%d", *bicepParam.MinValue))
	}
	if bicepParam.MaxValue != nil {
		sch.Maximum = json.Number(fmt.Sprintf("%d", *bicepParam.MaxValue))
	}

	return diags
}

func parseBoolParam(sch *schema.Schema, bicepParam *bicepParam) {
	sch.Type = "boolean"
	sch.Default = bicepParam.DefaultValue
}

func parseStringParam(sch *schema.Schema, bicepParam *bicepParam, secure bool, diags []result.Diagnostic) []result.Diagnostic {
	sch.Type = "string"
	sch.Default = bicepParam.DefaultValue

	if secure {
		sch.Format = "password"
	}

	allowedVals := bicepParam.AllowedValues
	if len(allowedVals) == 1 {
		assertedEnum, ok := allowedVals[0].([]interface{})
		if ok {
			sch.Enum = assertedEnum
		} else {
			sch.Comment = "Airlock Warning: unable to convert 'allowedValues' to enum"
			diags = append(diags, result.Diagnostic{
				Path:    sch.Title,
				Code:    "invalid_value",
				Message: fmt.Sprintf("unable to convert 'allowedValues' to enum in bicep param %s", sch.Title),
				Level:   result.Warning,
			})
		}
	}

	sch.MinLength = bicepParam.MinLength
	sch.MaxLength = bicepParam.MaxLength

	return diags
}

func parseArrayParam(sch *schema.Schema, bicepParam *bicepParam, diags []result.Diagnostic) []result.Diagnostic {
	sch.Type = "array"

	sch.MinItems = bicepParam.MinLength
	sch.MaxItems = bicepParam.MaxLength

	if bicepParam.DefaultValue != nil && len(bicepParam.DefaultValue.([]interface{})) != 0 {
		diags = parseArrayType(sch, bicepParam.DefaultValue.([]interface{}), diags)
	}
	return diags
}

func parseObjectParam(sch *schema.Schema, bicepParam *bicepParam, diags []result.Diagnostic) []result.Diagnostic {
	sch.Type = "object"

	if bicepParam.DefaultValue != nil && len(bicepParam.DefaultValue.(map[string]interface{})) > 1 {
		diags = parseObjectType(sch, bicepParam.DefaultValue.(map[string]interface{}), diags)
	}
	return diags
}

func parseObjectType(sch *schema.Schema, objValue map[string]interface{}, diags []result.Diagnostic) []result.Diagnostic {
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
			diags = parseArrayType(property, value.([]interface{}), diags)
		case reflect.Map:
			property.Type = "object"
			diags = parseObjectType(property, value.(map[string]interface{}), diags)
		default:
			sch.Comment = fmt.Sprintf("Airlock Warning: unknown type for field %s (%s)", name, reflect.TypeOf(value).Kind())
			diags = append(diags, result.Diagnostic{
				Path:    sch.Title,
				Code:    "unknown_type",
				Message: fmt.Sprintf("type of field %s is unsupported (%s)", sch.Title, reflect.TypeOf(value).Kind()),
				Level:   result.Warning,
			})
		}

		sch.Properties.Set(name, property)
		sch.Required = append(sch.Required, name)
		slices.Sort(sch.Required)
	}

	return diags
}

func parseArrayType(sch *schema.Schema, value []interface{}, diags []result.Diagnostic) []result.Diagnostic {
	if len(value) > 0 {
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
			diags = parseArrayType(items, elem.([]interface{}), diags)
		case reflect.Map:
			items.Type = "object"
			diags = parseObjectType(items, elem.(map[string]interface{}), diags)
		default:
			sch.Comment = fmt.Sprintf("Airlock Warning: unknown type (%s)", reflect.TypeOf(value).Kind())
			diags = append(diags, result.Diagnostic{
				Path:    sch.Title,
				Code:    "unknown_type",
				Message: fmt.Sprintf("type of field %s is unsupported (%s)", sch.Title, reflect.TypeOf(value).Kind()),
				Level:   result.Warning,
			})
		}

		sch.Items = items
	}
	return diags
}
