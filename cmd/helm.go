package cmd

import (
	"fmt"

	"github.com/massdriver-cloud/schema-generator/pkg/helm"
	"github.com/spf13/cobra"
)

var helmCmd = &cobra.Command{
	Use:  "helm",
	Args: cobra.ExactArgs(1),
	RunE: runHelm,
}

func init() {
	rootCmd.AddCommand(helmCmd)
}

func runHelm(cmd *cobra.Command, args []string) error {
	schema, err := helm.Run(args[0])
	if err != nil {
		return err
	}
	fmt.Println(schema)
	return nil
}
