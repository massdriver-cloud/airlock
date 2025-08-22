package cmd

import (
	"fmt"

	"github.com/massdriver-cloud/airlock/docs/helpdocs"
	"github.com/massdriver-cloud/airlock/pkg/helm"
	"github.com/spf13/cobra"
)

func NewCmdHelm() *cobra.Command {
	helmCmd := &cobra.Command{
		Use:   "helm",
		Short: "Helm (YAML) translations",
		Long:  helpdocs.MustRender("helm"),
	}

	// Import
	helmInputCmd := &cobra.Command{
		Use:   `input`,
		Short: "Ingest a helm values.yaml file and generate a JSON Schema",
		Args:  cobra.ExactArgs(1),
		Long:  helpdocs.MustRender("helm/input"),
		RunE:  runHelmInput,
	}

	helmCmd.AddCommand(helmInputCmd)

	return helmCmd
}

func runHelmInput(cmd *cobra.Command, args []string) error {
	result := helm.HelmToSchema(args[0])

	fmt.Print(result.PrettyDiags())
	fmt.Print(result.PrettySchema())

	return nil
}
