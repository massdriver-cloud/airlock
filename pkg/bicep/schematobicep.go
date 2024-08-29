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

func getBicepTypeFromInterface(interfaceType interface{}) string {
	switch reflect.TypeOf(interfaceType).Kind() {
	case reflect.String:
		return "string"
	case reflect.Float64:
		return "int"
	case reflect.Bool:
		return "bool"
	case reflect.Map:
		return "object"
	case reflect.Slice:
		return "array"
	}
	return ""
}

func renderBicep(val interface{}, prefix string) string {
	interfaceType := getBicepTypeFromInterface(val)
	switch interfaceType {
	case "string":
		return fmt.Sprintf("'%v'", val)
	case "int", "bool":
		return fmt.Sprintf("%v", val)
	case "array":
		return parseArray(val.([]interface{}), prefix)
	case "object":
		return parseObject(val.(map[string]interface{}), prefix)
	}
	return ""
}

func parseArray(arr []interface{}, prefix string) string {
	parsedArr := "[\n"
	fmt.Printf("unparsed array: %v\n", arr)
	prefix += "  "

	for _, v := range arr {
		parsedVal := renderBicep(v, prefix+"  ")
		parsedArr += fmt.Sprintf("%s%s", prefix, parsedVal) + "\n"
	}
	parsedArr += fmt.Sprintf("%s]", prefix[:len(prefix)-2])
	fmt.Printf("parsed array: %v\n", parsedArr)
	return parsedArr
}

func parseObject(obj map[string]interface{}, prefix string) string {
	fmt.Printf("unparsed obj: %v\n", obj)
	parsedObj := "{\n"
	prefix += "  "

	for k, v := range obj {
		parsedVal := renderBicep(v, prefix+"  ")
		parsedObj += fmt.Sprintf("%s%s: %s", prefix, k, parsedVal) + "\n"
	}

	parsedObj += fmt.Sprintf("%s}", prefix[:len(prefix)-2])
	fmt.Printf("parsed obj: %v\n", parsedObj)
	return parsedObj
}

func writeBicepParam(name string, sch *schema.Schema, buf *bytes.Buffer, bicepType string) {
	defVal := ""
	if sch.Default != nil {
		defVal = fmt.Sprintf(" = %v", renderBicep(sch.Default, ""))
	}

	buf.WriteString(fmt.Sprintf("param %s %s%v\n", name, bicepType, defVal))
}

func writeAllowedParams(sch *schema.Schema, buf *bytes.Buffer) {
	if sch.Enum != nil && len(sch.Enum) > 0 {
		buf.WriteString(fmt.Sprintf("@allowed(%v)\n", renderBicep(sch.Enum, "")))
	}
}

func writeDescription(sch *schema.Schema, buf *bytes.Buffer) {
	if sch.Description != "" {
		// decorators are in sys namespace. to avoid potential collision with other parameters named "description", we use "sys.description" instead of just "description" https://learn.microsoft.com/en-us/azure/azure-resource-manager/bicep/parameters#decorators
		buf.WriteString(fmt.Sprintf("@sys.description('%s')\n", sch.Description))
	}
}

func writeMinValue(sch *schema.Schema, buf *bytes.Buffer, bicepType string) {
	if bicepType == "int" {
		if sch.Minimum != "" {
			buf.WriteString(fmt.Sprintf("@minValue(%v)\n", sch.Minimum))
		}
	}
}

func writeMaxValue(sch *schema.Schema, buf *bytes.Buffer, bicepType string) {
	if bicepType == "int" {
		if sch.Maximum != "" {
			buf.WriteString(fmt.Sprintf("@maxValue(%v)\n", sch.Maximum))
		}
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
