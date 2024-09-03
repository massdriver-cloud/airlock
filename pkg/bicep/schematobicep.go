package bicep

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/massdriver-cloud/airlock/pkg/schema"
)

var indent string = "  "

func SchemaToBicep(in io.Reader) ([]byte, error) {
	inBytes, err := io.ReadAll(in)
	if err != nil {
		return nil, err
	}

	root := schema.Schema{}
	err = json.Unmarshal(inBytes, &root)
	if err != nil {
		return nil, err
	}

	content := bytes.NewBuffer(nil)

	flattenedProperties := schema.ExpandProperties(&root)
	for prop := flattenedProperties.Oldest(); prop != nil; prop = prop.Next() {
		err = createBicepParameter(prop.Key, prop.Value, content)
		if err != nil {
			return nil, err
		}
	}

	return content.Bytes(), nil
}

func createBicepParameter(name string, sch *schema.Schema, buf *bytes.Buffer) error {
	bicepType, err := getBicepTypeFromSchema(sch.Type)
	if err != nil {
		return err
	}

	writeDescription(sch, buf)
	writeAllowedParams(sch, buf)
	writeMinValue(sch, buf, bicepType)
	writeMaxValue(sch, buf, bicepType)
	writeMinLength(sch, buf, bicepType)
	writeMaxLength(sch, buf, bicepType)
	writeSecure(sch, buf, bicepType)
	writeBicepParam(name, sch, buf, bicepType)
	return nil
}

func writeBicepParam(name string, sch *schema.Schema, buf *bytes.Buffer, bicepType string) error {
	var defVal string

	if sch.Default != nil {
		renderedVal, err := renderBicep(sch.Default, "")
		if err != nil {
			return err
		}

		defVal = fmt.Sprintf(" = %s", renderedVal)
	}

	buf.WriteString(fmt.Sprintf("param %s %s%s\n", name, bicepType, defVal))
	return nil
}

func renderBicep(val interface{}, prefix string) (string, error) {
	interfaceType, err := getBicepTypeFromInterface(val)
	if err != nil {
		return "", err
	}

	switch interfaceType {
	case "string":
		return fmt.Sprintf("'%s'", val), nil
	case "int", "bool":
		return fmt.Sprintf("%v", val), nil
	case "array":
		assertedVal, asserArrErr := val.([]interface{})
		if asserArrErr != true {
			return "", fmt.Errorf("unable to convert value into array: %v", val)
		}

		return parseArray(assertedVal, prefix)
	case "object":
		assertedVal, asserObjErr := val.(map[string]interface{})
		if asserObjErr != true {
			return "", fmt.Errorf("unable to convert value into object: %v", val)
		}

		return parseObject(assertedVal, prefix)
	default:
		return "", err
	}
}

func getBicepTypeFromSchema(schemaType string) (string, error) {
	switch schemaType {
	case "string":
		return "string", nil
	case "integer", "number":
		return "int", nil
	case "boolean":
		return "bool", nil
	case "object", "":
		return "object", nil
	case "array":
		return "array", nil
	default:
		return "", errors.New("unknown type: " + schemaType)
	}
}

func getBicepTypeFromInterface(interfaceType interface{}) (string, error) {
	switch reflect.TypeOf(interfaceType).Kind() {
	case reflect.String:
		return "string", nil
	case reflect.Float64:
		return "int", nil
	case reflect.Bool:
		return "bool", nil
	case reflect.Map:
		return "object", nil
	case reflect.Slice:
		return "array", nil
	default:
		return "", errors.New("unknown type: " + reflect.TypeOf(interfaceType).Kind().String())
	}
}

func writeDescription(sch *schema.Schema, buf *bytes.Buffer) {
	if sch.Description != "" {
		// decorators are in sys namespace. to avoid potential collision with other parameters named "description", we use "sys.description" instead of just "description" https://learn.microsoft.com/en-us/azure/azure-resource-manager/bicep/parameters#decorators
		buf.WriteString(fmt.Sprintf("@sys.description('%s')\n", sch.Description))
	}
}

func writeAllowedParams(sch *schema.Schema, buf *bytes.Buffer) error {
	if sch.Enum != nil && len(sch.Enum) > 0 {
		renderedVal, err := renderBicep(sch.Enum, "")
		if err != nil {
			return err
		}

		buf.WriteString(fmt.Sprintf("@allowed(%s)\n", renderedVal))
	}
	return nil
}

func writeMinValue(sch *schema.Schema, buf *bytes.Buffer, bicepType string) {
	if bicepType == "int" && sch.Minimum != "" {
		// set this to %v because sch.Minimum uses json.Number type
		buf.WriteString(fmt.Sprintf("@minValue(%v)\n", sch.Minimum))
	}
}

func writeMaxValue(sch *schema.Schema, buf *bytes.Buffer, bicepType string) {
	if bicepType == "int" && sch.Maximum != "" {
		buf.WriteString(fmt.Sprintf("@maxValue(%v)\n", sch.Maximum))
	}
}

func writeMinLength(sch *schema.Schema, buf *bytes.Buffer, bicepType string) {
	switch bicepType {
	case "array":
		if sch.MinItems != nil {
			buf.WriteString(fmt.Sprintf("@minLength(%d)\n", *sch.MinItems))
		}
	case "string":
		if sch.MinLength != nil {
			buf.WriteString(fmt.Sprintf("@minLength(%d)\n", *sch.MinLength))
		}
	}
}

func writeMaxLength(sch *schema.Schema, buf *bytes.Buffer, bicepType string) {
	switch bicepType {
	case "array":
		if sch.MaxItems != nil {
			buf.WriteString(fmt.Sprintf("@maxLength(%d)\n", *sch.MaxItems))
		}
	case "string":
		if sch.MaxLength != nil {
			buf.WriteString(fmt.Sprintf("@maxLength(%d)\n", *sch.MaxLength))
		}
	}
}

func writeSecure(sch *schema.Schema, buf *bytes.Buffer, bicepType string) {
	if bicepType == "string" && sch.Format == "password" {
		buf.WriteString("@secure()\n")
	}
}

func parseArray(arr []interface{}, prefix string) (string, error) {
	parsedArr := "[\n"

	for _, v := range arr {
		renderedVal, err := renderBicep(v, prefix+indent)
		if err != nil {
			return "", err
		}

		parsedArr += fmt.Sprintf("%s%s", prefix+indent, renderedVal) + "\n"
	}

	parsedArr += fmt.Sprintf("%s]", prefix)
	return parsedArr, nil
}

func parseObject(obj map[string]interface{}, prefix string) (string, error) {
	parsedObj := "{\n"

	for k, v := range obj {
		renderedVal, err := renderBicep(v, prefix+indent)
		if err != nil {
			return "", err
		}

		parsedObj += fmt.Sprintf("%s%s: %s", prefix+indent, k, renderedVal) + "\n"
	}

	parsedObj += fmt.Sprintf("%s}", prefix)
	return parsedObj, nil
}
