package helm

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/iancoleman/orderedmap"
	"github.com/invopop/jsonschema"
	"gopkg.in/yaml.v3"
)

type nullError struct{}

func (e *nullError) Error() string {
	return "type is indeterminate (null)"
}

func Run(valuesPath string) (string, error) {
	valuesBytes, readErr := os.ReadFile(valuesPath)
	if readErr != nil {
		return "", readErr
	}

	valuesDocument := yaml.Node{}

	unmarshalErr := yaml.Unmarshal(valuesBytes, &valuesDocument)
	if unmarshalErr != nil {
		return "", unmarshalErr
	}

	// the top level node is a document node. We need to go one layer
	// deeper to get the actual yaml content
	params := new(jsonschema.Schema)
	err := parseMapNode(params, valuesDocument.Content[0])
	if err != nil {
		return "", err
	}

	bytes, err := json.MarshalIndent(params, "", "  ")
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func parseNameNode(schema *jsonschema.Schema, node *yaml.Node) error {
	schema.Title = node.Value

	description := strings.TrimLeft(node.HeadComment, "# \t")
	if len(description) > 0 {
		schema.Description = description
	}

	return nil
}

func parseValueNode(schema *jsonschema.Schema, node *yaml.Node) error {
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

func nodeToProperty(name, value *yaml.Node) (*jsonschema.Schema, error) {
	schema := new(jsonschema.Schema)

	if err := parseNameNode(schema, name); err != nil {
		return nil, err
	}

	err := parseValueNode(schema, value)
	if err != nil {
		if _, ok := err.(*nullError); ok {
			fmt.Printf("Warning: Skipping field %s. Reason: %v\n", schema.Title, err)
			return nil, nil
		} else {
			return nil, err
		}
	}

	return schema, nil
}

func parseStringNode(schema *jsonschema.Schema, node *yaml.Node) error {
	schema.Type = "string"
	schema.Default = node.Value
	return nil
}

func parseIntegerNode(schema *jsonschema.Schema, node *yaml.Node) error {
	schema.Type = "integer"
	def, err := strconv.Atoi(node.Value)
	if err != nil {
		return err
	}
	schema.Default = def
	return nil
}

func parseFloatNode(schema *jsonschema.Schema, node *yaml.Node) error {
	schema.Type = "number"
	def, err := strconv.ParseFloat(node.Value, 64)
	if err != nil {
		return err
	}
	schema.Default = def
	return nil
}

func parseBooleanNode(schema *jsonschema.Schema, node *yaml.Node) error {
	schema.Type = "boolean"
	def, err := strconv.ParseBool(node.Value)
	if err != nil {
		return err
	}
	schema.Default = def
	return nil
}

func parseMapNode(schema *jsonschema.Schema, node *yaml.Node) error {
	schema.Type = "object"
	schema.Properties = orderedmap.New()

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
			schema.Properties.Set(nameNode.Value, property)
			schema.Required = append(schema.Required, nameNode.Value)
		}
	}

	return nil
}

func parseArrayNode(schema *jsonschema.Schema, node *yaml.Node) error {
	schema.Type = "array"

	if len(node.Content) == 0 {
		return fmt.Errorf("error: cannot infer element type in array %s. Arrays cannot be empty or the type is ambiguous", schema.Title)
	}
	schema.Items = new(jsonschema.Schema)
	parseValueNode(schema.Items, node.Content[0])

	// Set the default back to nil since we don't want to default all items to the first type in the list
	node.Decode(&schema.Default)

	return nil
}
