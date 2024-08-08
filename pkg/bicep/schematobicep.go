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
		enumbytes, err := json.MarshalIndent(sch.Enum, "", "    ")
		if err != nil {
			return err
		}

		enumstring := string(enumbytes)
		r := strings.NewReplacer(`"`, `'`, ",", "")

		buf.WriteString(fmt.Sprintf("@allowed(%v)\n", r.Replace(enumstring)))
	}

	return nil
}

// decorators are in sys namespace. to avoid potential collision with other parameters named "description"
func declareDescription(sch *schema.Schema, buf *bytes.Buffer) {
	if sch.Description != "" {
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

func declareMinLength(sch *schema.Schema, buf *bytes.Buffer) {
	if sch.MinLength != nil {
		buf.WriteString(fmt.Sprintf("@minLength(%d)\n", sch.MinLength))
	}
}

func declareMaxLength(sch *schema.Schema, buf *bytes.Buffer) {
	if sch.MaxLength != nil {
		buf.WriteString(fmt.Sprintf("@maxLength(%d)\n", sch.MaxLength))
	}
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

		if bicepType == "string" {
			buf.WriteString(fmt.Sprintf("param %s %s = '%s'\n", name, bicepType, sch.Default))
		}

		if bicepType == "int" {
			buf.WriteString(fmt.Sprintf("param %s %s = %v\n", name, bicepType, sch.Default))
		}

		if bicepType == "bool" {
			buf.WriteString(fmt.Sprintf("param %s %s = %t\n", name, bicepType, sch.Default))
		}

		if bicepType == "array" {
			buf.WriteString(fmt.Sprintf("param %s %s = %v\n", name, bicepType, sch.Default))
		}

		if bicepType == "object" {
			buf.WriteString(fmt.Sprintf("param %s %s = %v\n", name, bicepType, sch.Default))
		}
	}
	return nil
}
