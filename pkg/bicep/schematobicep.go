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
	declareAllowed(sch, buf)
	declareDescription(sch, buf)
	declareMinValue(sch, buf)
	declareMaxValue(sch, buf)
	declareMinLength(sch, buf)
	declareMaxLength(sch, buf)
	declareSecure(sch, buf)
	declareParameter(name, sch, buf)
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

func declareParameter(name string, sch *schema.Schema, buf *bytes.Buffer) error {
	bicepType, err := getBicepType(sch.Type)
	if err != nil {
		return err
	}

	if sch.Default != nil {
		err = declareDefault(name, sch, buf)
		if err != nil {
			return err
		}
		return nil
	} else {
		buf.WriteString(fmt.Sprintf("param %s %s\n", name, bicepType))
	}

	return nil
}

func declareAllowed(sch *schema.Schema, buf *bytes.Buffer) error {
	if sch.Enum != nil && len(sch.Enum) > 0 {
		bicepType, err := getBicepType(sch.Type)
		if err != nil {
			return err
		}

		enumBytes, err := json.MarshalIndent(sch.Enum, "", "    ")
		if err != nil {
			return err
		}

		enumString := string(enumBytes)
		r := strings.NewReplacer(`"`, `'`, ",", "")
		cleanString := r.Replace(enumString)

		if bicepType == "object" {
			splitString := strings.Split(cleanString, " ")
			joinList := []string{}
			for _, d := range splitString {
				if strings.Contains(d, ":") {
					d = strings.ReplaceAll(d, `'`, "")
				}
				joinList = append(joinList, d)
			}
			bicepObj := strings.Join(joinList, " ")
			buf.WriteString(fmt.Sprintf("@allowed(%v)\n", bicepObj))
		} else {
			buf.WriteString(fmt.Sprintf("@allowed(%v)\n", cleanString))
		}
	}

	return nil
}

func declareDescription(sch *schema.Schema, buf *bytes.Buffer) {
	if sch.Description != "" {
		// decorators are in sys namespace. to avoid potential collision with other parameters named "description", we use "sys.description" instead of just "description" https://learn.microsoft.com/en-us/azure/azure-resource-manager/bicep/parameters#decorators
		buf.WriteString(fmt.Sprintf("@sys.description('%s')\n", sch.Description))
	}
}

func declareMinValue(sch *schema.Schema, buf *bytes.Buffer) {
	if sch.Minimum != "" {
		buf.WriteString(fmt.Sprintf("@minValue(%v)\n", sch.Minimum))
	}
}

func declareMaxValue(sch *schema.Schema, buf *bytes.Buffer) {
	if sch.Maximum != "" {
		buf.WriteString(fmt.Sprintf("@maxValue(%v)\n", sch.Maximum))
	}
}

func declareMinLength(sch *schema.Schema, buf *bytes.Buffer) error {
	bicepType, err := getBicepType(sch.Type)
	if err != nil {
		return err
	}

	if bicepType == "array" {
		if sch.MinItems != nil {
			buf.WriteString(fmt.Sprintf("@minLength(%d)\n", *sch.MinItems))
		}
		return nil
	}

	if sch.MinLength != nil {
		buf.WriteString(fmt.Sprintf("@minLength(%d)\n", *sch.MinLength))
	}

	return nil
}

func declareMaxLength(sch *schema.Schema, buf *bytes.Buffer) error {
	bicepType, err := getBicepType(sch.Type)
	if err != nil {
		return err
	}

	if bicepType == "array" {
		if sch.MaxItems != nil {
			buf.WriteString(fmt.Sprintf("@maxLength(%d)\n", *sch.MaxItems))
		}
		return nil
	}

	if sch.MaxLength != nil {
		buf.WriteString(fmt.Sprintf("@maxLength(%d)\n", *sch.MaxLength))
	}

	return nil
}

func declareSecure(sch *schema.Schema, buf *bytes.Buffer) {
	if sch.Format == "password" {
		buf.WriteString("@secure()\n")
	}
}

func declareDefault(name string, sch *schema.Schema, buf *bytes.Buffer) error {
	if sch.Default != nil {
		bicepType, err := getBicepType(sch.Type)
		if err != nil {
			return err
		}

		switch bicepType {
		case "string":
			buf.WriteString(fmt.Sprintf("param %s %s = '%s'\n", name, bicepType, sch.Default))
		case "int":
			buf.WriteString(fmt.Sprintf("param %s %s = %v\n", name, bicepType, sch.Default))
		case "bool":
			buf.WriteString(fmt.Sprintf("param %s %s = %t\n", name, bicepType, sch.Default))
		case "array":
			declareDefaultArray(name, sch, buf)
		case "object":
			declareDefaultObject(name, sch, buf)
		}
	}
	return nil
}

func declareDefaultArray(name string, sch *schema.Schema, buf *bytes.Buffer) error {
	bicepType, err := getBicepType(sch.Type)
	if err != nil {
		return err
	}

	defBytes, err := json.MarshalIndent(sch.Default.([]interface{}), "", "    ")
	if err != nil {
		return err
	}

	defString := string(defBytes)
	r := strings.NewReplacer(`"`, `'`, ",", "")

	buf.WriteString(fmt.Sprintf("param %s %s = %v\n", name, bicepType, r.Replace(defString)))
	return nil
}

func declareDefaultObject(name string, sch *schema.Schema, buf *bytes.Buffer) error {
	bicepType, err := getBicepType(sch.Type)
	if err != nil {
		return err
	}

	defBytes, err := json.MarshalIndent(sch.Default, "", "    ")
	if err != nil {
		return err
	}

	r := strings.NewReplacer(`"`, `'`, ",", "")
	byteToStr := string(defBytes)
	cleanString := r.Replace(byteToStr)

	splitString := strings.Split(cleanString, " ")
	joinList := []string{}
	for _, d := range splitString {
		if strings.Contains(d, ":") {
			d = strings.ReplaceAll(d, `'`, "")
		}
		joinList = append(joinList, d)
	}
	bicepObj := strings.Join(joinList, " ")

	buf.WriteString(fmt.Sprintf("param %s %s = %v\n", name, bicepType, bicepObj))
	return nil
}
