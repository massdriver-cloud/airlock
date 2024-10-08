package helm

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/massdriver-cloud/airlock/pkg/schema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	yaml "gopkg.in/yaml.v3"
)

type nullError struct{}

func (e *nullError) Error() string {
	return "type is indeterminate (null)"
}

func HelmToSchema(valuesPath string) (*schema.Schema, error) {
	valuesBytes, readErr := os.ReadFile(valuesPath)
	if readErr != nil {
		return nil, readErr
	}

	valuesDocument := yaml.Node{}

	unmarshalErr := yaml.Unmarshal(valuesBytes, &valuesDocument)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	// the top level node is a document node. We need to go one layer
	// deeper to get the actual yaml content
	sch := new(schema.Schema)
	err := parseMapNode(sch, valuesDocument.Content[0])
	if err != nil {
		return nil, err
	}

	return sch, nil
}

func parseNameNode(schema *schema.Schema, node *yaml.Node) {
	schema.Title = node.Value

	description := strings.TrimLeft(node.HeadComment, "# \t")
	if len(description) > 0 {
		schema.Description = description
	}
}

func parseValueNode(schema *schema.Schema, node *yaml.Node) error {
	switch node.Tag {
	case "!!str":
		return parseStringNode(schema, node)
	case "!!int":
		return parseIntegerNode(schema, node)
	case "!!float":
		return parseFloatNode(schema, node)
	case "!!bool":
		return parseBooleanNode(schema, node)
	case "!!map":
		return parseMapNode(schema, node)
	case "!!seq":
		return parseArrayNode(schema, node)
	case "!!null":
		return &nullError{}
	default:
		return fmt.Errorf("unrecognized tag %s", node.Tag)
	}
}

func nodeToProperty(name, value *yaml.Node) (*schema.Schema, error) {
	sch := new(schema.Schema)

	parseNameNode(sch, name)

	err := parseValueNode(sch, value)
	if err != nil {
		//nolint:errorlint
		if _, ok := err.(*nullError); ok {
			fmt.Printf("warning: skipping field %s\n reason: %v\n", sch.Title, err)
			//nolint:nilnil
			return nil, nil
		}
		return nil, err
	}

	return sch, nil
}

func parseStringNode(sch *schema.Schema, node *yaml.Node) error {
	sch.Type = "string"
	sch.Default = node.Value
	return nil
}

func parseIntegerNode(sch *schema.Schema, node *yaml.Node) error {
	sch.Type = "integer"
	def, err := strconv.Atoi(node.Value)
	if err != nil {
		return err
	}
	sch.Default = def
	return nil
}

func parseFloatNode(sch *schema.Schema, node *yaml.Node) error {
	sch.Type = "number"
	def, err := strconv.ParseFloat(node.Value, 64)
	if err != nil {
		return err
	}
	sch.Default = def
	return nil
}

func parseBooleanNode(sch *schema.Schema, node *yaml.Node) error {
	sch.Type = "boolean"
	def, err := strconv.ParseBool(node.Value)
	if err != nil {
		return err
	}
	sch.Default = def
	return nil
}

func parseMapNode(sch *schema.Schema, node *yaml.Node) error {
	sch.Type = "object"
	sch.Properties = orderedmap.New[string, *schema.Schema]()

	nodes := node.Content
	// Nodes come in twos - the first is the name, the second is the value
	for index := 0; index < len(nodes); index += 2 {
		nameNode := nodes[index]
		valueNode := nodes[index+1]
		property, err := nodeToProperty(nameNode, valueNode)
		if err != nil {
			return err
		}
		if property != nil {
			sch.Properties.Set(nameNode.Value, property)
			sch.Required = append(sch.Required, nameNode.Value)
		}
	}

	return nil
}

func parseArrayNode(sch *schema.Schema, node *yaml.Node) error {
	sch.Type = "array"

	if len(node.Content) == 0 {
		return &nullError{}
	}
	sch.Items = new(schema.Schema)
	err := parseValueNode(sch.Items, node.Content[0])
	if err != nil {
		return err
	}

	// Set the default back to nil since we don't want to default all items to the first type in the list
	err = node.Decode(&sch.Default)
	if err != nil {
		return err
	}

	return nil
}
