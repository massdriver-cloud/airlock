package helm

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/massdriver-cloud/airlock/pkg/result"
	"github.com/massdriver-cloud/airlock/pkg/schema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	yaml "gopkg.in/yaml.v3"
)

func HelmToSchema(valuesPath string) result.SchemaResult {
	valuesBytes, readErr := os.ReadFile(valuesPath)
	if readErr != nil {
		return result.SchemaResult{
			Schema: nil,
			Diags: []result.Diagnostic{
				{
					Path:    valuesPath,
					Code:    "file_read_error",
					Message: fmt.Sprintf("failed to read values file: %s", readErr),
					Level:   result.Error,
				},
			},
		}
	}

	valuesDocument := yaml.Node{}

	unmarshalErr := yaml.Unmarshal(valuesBytes, &valuesDocument)
	if unmarshalErr != nil {
		return result.SchemaResult{
			Schema: nil,
			Diags: []result.Diagnostic{
				{
					Path:    valuesPath,
					Code:    "yaml_unmarshal_error",
					Message: fmt.Sprintf("failed to unmarshal values file: %s", unmarshalErr),
					Level:   result.Error,
				},
			},
		}
	}

	sch := new(schema.Schema)
	result := result.SchemaResult{
		Schema: sch,
		Diags:  []result.Diagnostic{},
	}

	// the top level node is a document node. We need to go one layer
	// deeper to get the actual yaml content
	result.Diags = parseMapNode(sch, valuesDocument.Content[0], result.Diags)

	return result
}

func parseNameNode(schema *schema.Schema, node *yaml.Node) {
	schema.Title = node.Value

	description := strings.TrimLeft(node.HeadComment, "# \t")
	if len(description) > 0 {
		schema.Description = description
	}
}

func parseValueNode(schema *schema.Schema, node *yaml.Node, diags []result.Diagnostic) []result.Diagnostic {
	switch node.Tag {
	case "!!str":
		parseStringNode(schema, node)
	case "!!int":
		return parseIntegerNode(schema, node, diags)
	case "!!float":
		return parseFloatNode(schema, node, diags)
	case "!!bool":
		return parseBooleanNode(schema, node, diags)
	case "!!map":
		return parseMapNode(schema, node, diags)
	case "!!seq":
		return parseArrayNode(schema, node, diags)
	case "!!null":
		schema.Comment = "Airlock Warning: unknown type from null value"
		return append(diags, result.Diagnostic{
			Path:    schema.Title,
			Code:    "unknown_type",
			Message: fmt.Sprintf("type of field %s is indeterminate (null)", schema.Title),
			Level:   result.Warning,
		})
	default:
		schema.Comment = fmt.Sprintf("Airlock Warning: unknown type %s", node.Tag)
		return append(diags, result.Diagnostic{
			Path:    schema.Title,
			Code:    "unknown_type",
			Message: fmt.Sprintf("type of field %s is unsupported (%s)", schema.Title, node.Tag),
			Level:   result.Warning,
		})
	}
	return diags
}

func nodeToProperty(sch *schema.Schema, name, value *yaml.Node, diags []result.Diagnostic) []result.Diagnostic {
	parseNameNode(sch, name)

	diags = parseValueNode(sch, value, diags)

	return diags
}

func parseStringNode(sch *schema.Schema, node *yaml.Node) {
	sch.Type = "string"
	sch.Default = node.Value
}

func parseIntegerNode(sch *schema.Schema, node *yaml.Node, diags []result.Diagnostic) []result.Diagnostic {
	sch.Type = "integer"
	def, err := strconv.Atoi(node.Value)
	if err != nil {
		return append(diags, result.Diagnostic{
			Path:    node.Value,
			Code:    "invalid_value",
			Message: fmt.Sprintf("failed to parse integer: %s", err),
			Level:   result.Error,
		})
	}
	sch.Default = def
	return diags
}

func parseFloatNode(sch *schema.Schema, node *yaml.Node, diags []result.Diagnostic) []result.Diagnostic {
	sch.Type = "number"
	def, err := strconv.ParseFloat(node.Value, 64)
	if err != nil {
		return append(diags, result.Diagnostic{
			Path:    node.Value,
			Code:    "invalid_value",
			Message: fmt.Sprintf("failed to parse float: %s", err),
			Level:   result.Error,
		})
	}
	sch.Default = def
	return diags
}

func parseBooleanNode(sch *schema.Schema, node *yaml.Node, diags []result.Diagnostic) []result.Diagnostic {
	sch.Type = "boolean"
	def, err := strconv.ParseBool(node.Value)
	if err != nil {
		return append(diags, result.Diagnostic{
			Path:    node.Value,
			Code:    "invalid_value",
			Message: fmt.Sprintf("failed to parse boolean: %s", err),
			Level:   result.Error,
		})
	}
	sch.Default = def
	return diags
}

func parseMapNode(sch *schema.Schema, node *yaml.Node, diags []result.Diagnostic) []result.Diagnostic {
	sch.Type = "object"
	sch.Properties = orderedmap.New[string, *schema.Schema]()

	nodes := node.Content
	// Nodes come in twos - the first is the name, the second is the value
	for index := 0; index < len(nodes); index += 2 {
		nameNode := nodes[index]
		valueNode := nodes[index+1]

		property := new(schema.Schema)
		diags = nodeToProperty(property, nameNode, valueNode, diags)

		sch.Properties.Set(nameNode.Value, property)
		sch.Required = append(sch.Required, nameNode.Value)
	}

	return diags
}

func parseArrayNode(sch *schema.Schema, node *yaml.Node, diags []result.Diagnostic) []result.Diagnostic {
	sch.Type = "array"

	sch.Items = new(schema.Schema)

	if len(node.Content) == 0 {
		sch.Items.Comment = "Airlock Warning: unknown type from empty array"
		return append(diags, result.Diagnostic{
			Path:    sch.Title,
			Code:    "unknown_type",
			Message: fmt.Sprintf("array %s is empty so it's type is unknown", sch.Title),
			Level:   result.Warning,
		})
	}

	diags = parseValueNode(sch.Items, node.Content[0], diags)

	// Set the default back to nil since we don't want to default all items to the first type in the list
	decodeErr := node.Decode(&sch.Default)
	if decodeErr != nil {
		return append(diags, result.Diagnostic{
			Path:    sch.Title,
			Code:    "invalid_type",
			Message: fmt.Sprintf("failed to decode array default: %s", decodeErr),
			Level:   result.Error,
		})
	}

	return diags
}
