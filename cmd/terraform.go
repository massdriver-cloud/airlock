package cmd

import (
	"fmt"

	"github.com/massdriver-cloud/airlock/pkg/terraform"
	"github.com/spf13/cobra"
)

var terraformCmd = &cobra.Command{
	Use:  "terraform",
	Args: cobra.ExactArgs(1),
	RunE: runTerraform,
}

func init() {
	rootCmd.AddCommand(terraformCmd)
}

func runTerraform(cmd *cobra.Command, args []string) error {
	schema, err := terraform.Run(args[0])
	if err != nil {
		return err
	}
	fmt.Println(schema)
	return nil
}
