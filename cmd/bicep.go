package cmd

import (
	"fmt"

	"github.com/massdriver-cloud/airlock/pkg/bicep"
	"github.com/spf13/cobra"
)

func NewCmdBicep() *cobra.Command {
	bicepCmd := &cobra.Command{
		Use:   "bicep",
		Short: "Bicep translations",
		Long:  helpdocs.MustRender("bicep"),
	}

	// Import
	bicepInputCmd := &cobra.Command{
		Use:   `input`,
		Short: "Ingest a bicep template file and generate a JSON Schema",
		Args:  cobra.ExactArgs(1),
		Long:  helpdocs.MustRender("bicep/input"),
		RunE: runBicepInput,
	}

	bicepCmd.AddCommand(bicepInputCmd)

	return bicepCmd
}

func runBicepInput(cmd *cobra.Command, args []string) error {
	schema, err := bicep.BicepToSchema(args[0])
	if err != nil {
		return err
	}
	fmt.Println(schema)
	return nil
}
