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

	writeParamDescription(sch, buf)
	writeParamMinValue(sch, buf, bicepType)
	writeParamMaxValue(sch, buf, bicepType)
	writeParamMinLength(sch, buf, bicepType)
	writeParamMaxLength(sch, buf, bicepType)
	writeSecureParam(sch, buf, bicepType)
	writeBicepParameter(name, sch, buf, bicepType)
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

func writeBicepParameter(name string, sch *schema.Schema, buf *bytes.Buffer, bicepType string) {
	defaultVal := writeDefaultParam(name, sch, buf, bicepType)
	buf.WriteString(fmt.Sprintf("param %s %s%v\n", name, bicepType, defaultVal))
}

func writeAllowedParams(sch *schema.Schema, buf *bytes.Buffer, bicepType string) error {
	if sch.Enum != nil && len(sch.Enum) > 0 {
		enumBytes, err := json.MarshalIndent(sch.Enum, "", "    ")
		if err != nil {
			return err
		}

		enumString := string(enumBytes)
		r := strings.NewReplacer(`"`, `'`, ",", "")
		cleanString := r.Replace(enumString)

		if bicepType == "object" {
			parseParamObject(cleanString, sch, buf, bicepType)
			// add recursive call here
		} else {
			buf.WriteString(fmt.Sprintf("@allowed(%v)\n", cleanString))
		}
	}

	return nil
}

func writeParamDescription(sch *schema.Schema, buf *bytes.Buffer) {
	if sch.Description != "" {
		// decorators are in sys namespace. to avoid potential collision with other parameters named "description", we use "sys.description" instead of just "description" https://learn.microsoft.com/en-us/azure/azure-resource-manager/bicep/parameters#decorators
		buf.WriteString(fmt.Sprintf("@sys.description('%s')\n", sch.Description))
	}
}

func writeParamMinValue(sch *schema.Schema, buf *bytes.Buffer, bicepType string) {
	if bicepType == "int" {
		if sch.Minimum != "" {
			buf.WriteString(fmt.Sprintf("@minValue(%v)\n", sch.Minimum))
		}
	}
}

func writeParamMaxValue(sch *schema.Schema, buf *bytes.Buffer, bicepType string) {
	if bicepType == "int" {
		if sch.Maximum != "" {
			buf.WriteString(fmt.Sprintf("@maxValue(%v)\n", sch.Maximum))
		}
	}
}

func writeParamMinLength(sch *schema.Schema, buf *bytes.Buffer, bicepType string) {
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

func writeParamMaxLength(sch *schema.Schema, buf *bytes.Buffer, bicepType string) {
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

func writeSecureParam(sch *schema.Schema, buf *bytes.Buffer, bicepType string) {
	if (bicepType == "string" || bicepType == "object") && sch.Format == "password" {
		buf.WriteString("@secure()\n")
	}
}

func writeDefaultParam(name string, sch *schema.Schema, buf *bytes.Buffer, bicepType string) string {
	if sch.Default != nil {
		switch bicepType {
		case "string":
			return fmt.Sprintf(" = '%v'", sch.Default)
		case "int", "bool":
			return fmt.Sprintf(" = %v", sch.Default)
		case "array":
			parseParamArray(name, sch, buf, bicepType)
		case "object":
			parseParamObject(name, sch, buf, bicepType)
		}
	}
	return ""
}

func parseParamArray(name string, sch *schema.Schema, buf *bytes.Buffer, bicepType string) (string, error) {
	defBytes, err := json.MarshalIndent(sch.Default.([]interface{}), "", "    ")
	if err != nil {
		return "", err
	}

	defString := string(defBytes)
	r := strings.NewReplacer(`"`, `'`, ",", "")

	// array of objects result?

	return fmt.Sprintf("%v", r.Replace(defString)), nil
}

func parseParamObject(name string, sch *schema.Schema, buf *bytes.Buffer, bicepType string) (string, error) {
	defBytes, err := json.MarshalIndent(sch.Default, "", "    ")
	if err != nil {
		return "", err
	}

	r := strings.NewReplacer(`"`, `'`, ",", "")
	byteToStr := string(defBytes)
	cleanString := r.Replace(byteToStr)

	// what if default value in object is a string that has a space?
	splitString := strings.Split(cleanString, " ")
	joinList := []string{}
	for _, d := range splitString {
		// explain this
		if strings.Contains(d, ":") {
			d = strings.ReplaceAll(d, `'`, "")
		}
		joinList = append(joinList, d)
	}
	bicepObj := strings.Join(joinList, " ")

	return fmt.Sprintf("%v", bicepObj), nil
}

// func object2(str string, sch *schema.Schema, buf *bytes.Buffer) string {
// 	// different than declareDefaultObject?
// 	// find library?
// 	splitString := strings.Split(str, " ")
// 	joinList := []string{}
// 	for _, d := range splitString {
// 		if strings.Contains(d, ":") {
// 			d = strings.ReplaceAll(d, `'`, "")
// 		}
// 		joinList = append(joinList, d)
// 	}
// 	bicepObj := strings.Join(joinList, " ")
// 	return fmt.Sprintf("@allowed(%v)\n", bicepObj)
// }
