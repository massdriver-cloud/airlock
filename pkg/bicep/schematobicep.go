package bicep

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

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

	if sch.Enum != nil {
		err = declareAllowed(sch, buf)
		if err != nil {
			return err
		}
	}

	buf.WriteString(fmt.Sprintf("param %s %s\n", name, bicepType))
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

func declareAllowed(sch *schema.Schema, buf *bytes.Buffer) error {
	enums, err := getEnums(sch.Enum)
	if err != nil {
		return err
	}

	buf.WriteString(fmt.Sprintf("@allowed(%s)\n", enums))
	return nil
}

func getEnums(schemaEnums []interface{}) (string, error) {
	enums := []string{"\n"}
	for _, enum := range schemaEnums {
		enums = append(enums, fmt.Sprintf("   '%s'\n", enum))
	}
	return fmt.Sprintf("%s", enums), nil
}
