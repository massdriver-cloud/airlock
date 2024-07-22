package cmd

import (
	"fmt"

	"github.com/massdriver-cloud/airlock/pkg/bicep"
	"github.com/spf13/cobra"
)

func NewCmdBicep() *cobra.Command {
	helmCmd := &cobra.Command{
		Use:   "bicep",
		Short: "Bicep translations",
		//Long:  helpdocs.MustRender("artifact"),
	}

	// Import
	helmInputCmd := &cobra.Command{
		Use:   `input`,
		Short: "Ingest a bicep template file and generate a JSON Schema",
		Args:  cobra.ExactArgs(1),
		// Long:  helpdocs.MustRender("helm/input"),
		RunE: runBicepInput,
	}

	helmCmd.AddCommand(helmInputCmd)

	return helmCmd
}

func runBicepInput(cmd *cobra.Command, args []string) error {
	schema, err := bicep.BicepToSchema(args[0])
	if err != nil {
		return err
	}
	fmt.Println(schema)
	return nil
}
