package result

import (
	"encoding/json"
	"fmt"

	"github.com/massdriver-cloud/airlock/pkg/prettylogs"
)

func (result *SchemaResult) PrettyDiags() string {
	output := ""
	for _, diag := range result.Diags {
		levelString := prettylogs.Orange("WARNING")
		if diag.Level == Error {
			levelString = prettylogs.Red("ERROR")
		}
		output += fmt.Sprintf("Airlock %s: %s\n", levelString, diag.Message)
	}
	return output
}

func (result *SchemaResult) PrettySchema() string {
	if result.Schema == nil {
		return "No schema available"
	}
	bytes, err := json.MarshalIndent(result.Schema, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error marshaling schema: %s", err)
	}
	return string(bytes)
}
