package result

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (result *SchemaResult) PrettyDiags() string {
	warningString := lipgloss.NewStyle().SetString("WARNING").Foreground(lipgloss.Color("#FFA500"))
	errorString := lipgloss.NewStyle().SetString("ERROR").Foreground(lipgloss.Color("#FF0000"))

	output := ""
	for _, diag := range result.Diags {
		levelString := warningString
		if diag.Level == Error {
			levelString = errorString
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
