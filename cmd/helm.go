package cmd

import (
	"encoding/json"
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
	schema, err := helm.HelmToSchema(args[0])
	if err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(bytes))
	return nil
}
