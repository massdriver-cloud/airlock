package bicep

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

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
	bicepType, err := getBicepType(sch.Type)
	if err != nil {
		return err
	}

	writeDescription(sch, buf)
	writeAllowedParams(sch, buf, bicepType)
	writeMinValue(sch, buf, bicepType)
	writeMaxValue(sch, buf, bicepType)
	writeMinLength(sch, buf, bicepType)
	writeMaxLength(sch, buf, bicepType)
	writeSecure(sch, buf, bicepType)
	writeBicepParam(name, sch, buf, bicepType)
	return nil
}

func getBicepType(schemaType string) (string, error) {
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

func renderBicep(val interface{}, bicepType string) string {
	switch bicepType {
	case "string":
		return fmt.Sprintf("'%v'", val)
	case "int", "bool":
		return fmt.Sprintf("%v", val)
	case "array":
		parseArray(val.([]interface{}))
	case "object":
		parseObject(val)
	}
	return ""
}

func parseArray(arr []interface{}) (string, error) {
	defBytes, err := json.MarshalIndent(arr, "", "  ")
	if err != nil {
		return "", err
	}

	defString := string(defBytes)
	r := strings.NewReplacer(`"`, `'`, ",", "")

	// call converValueToBicep?
	// handle new line, commas, etc. for arrays

	return fmt.Sprintf("%v", r.Replace(defString)), nil
}

func parseObject(obj interface{}) (string, error) {
	defBytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", err
	}

	r := strings.NewReplacer(`"`, `'`, ",", "")
	byteToStr := string(defBytes)
	cleanString := r.Replace(byteToStr)

	splitString := strings.Split(cleanString, " ")
	joinList := []string{}
	for _, d := range splitString {
		// come back to this, need to find library maybe?
		if strings.Contains(d, ":") {
			d = strings.ReplaceAll(d, `'`, "")
		}
		joinList = append(joinList, d)
	}
	bicepObj := strings.Join(joinList, " ")

	return fmt.Sprintf("%v", bicepObj), nil
}

func writeBicepParam(name string, sch *schema.Schema, buf *bytes.Buffer, bicepType string) {
	defVal := ""
	if sch.Default != nil {
		renderedVal := renderBicep(sch.Default, bicepType)

		if bicepType == "object" {
			renderedVal, _ = parseObject(sch.Default)
		}

		if bicepType == "array" {
			renderedVal, _ = parseArray(sch.Default.([]interface{}))
		}

		defVal = fmt.Sprintf(" = %v", renderedVal)
	}

	buf.WriteString(fmt.Sprintf("param %s %s%v\n", name, bicepType, defVal))
}

func writeAllowedParams(sch *schema.Schema, buf *bytes.Buffer, bicepType string) error {
	if sch.Enum != nil && len(sch.Enum) > 0 {
		parsedVal, err := parseArray(sch.Enum)
		if err != nil {
			return err
		}

		if bicepType == "object" {
			parsedVal, err = parseObject(sch.Enum)
			if err != nil {
				return err
			}
		}
		buf.WriteString(fmt.Sprintf("@allowed(%v)\n", parsedVal))
	}

	return nil
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
