package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/massdriver-cloud/airlock/docs/helpdocs"
	"github.com/massdriver-cloud/airlock/pkg/terraform"
	"github.com/spf13/cobra"
)

func NewCmdTerraform() *cobra.Command {
	terraformCmd := &cobra.Command{
		Use:     "terraform",
		Short:   "Terraform (HCL) translations",
		Aliases: []string{"tf", "terraform", "opentofu"},
		Long:    helpdocs.MustRender("terraform"),
	}

	// Input
	terraformInputCmd := &cobra.Command{
		Use:   `input`,
		Short: "Ingest a terraform module and generate a JSON Schema from the variables",
		Args:  cobra.ExactArgs(1),
		Long:  helpdocs.MustRender("terraform/input"),
		RunE:  runTerraformInput,
	}

	// oputput
	terraformOutputCmd := &cobra.Command{
		Use:   `output`,
		Short: "Output a terraform variables specification from a JSON schemea document",
		Args:  cobra.ExactArgs(1),
		Long:  helpdocs.MustRender("terraform/output"),
		RunE:  runTerraformOutput,
	}

	terraformCmd.AddCommand(terraformInputCmd)
	terraformCmd.AddCommand(terraformOutputCmd)

	return terraformCmd
}

func runTerraformInput(cmd *cobra.Command, args []string) error {
	schema, err := terraform.TfToSchema(args[0])
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

func runTerraformOutput(cmd *cobra.Command, args []string) error {
	schemaPath := args[0]

	var err error
	var in *os.File
	if schemaPath == "-" {
		in = os.Stdin
	} else {
		in, err = os.Open(schemaPath)
		if err != nil {
			return err
		}
		defer in.Close()
	}

	bytes, err := terraform.SchemaToTf(in)
	if err != nil {
		return err
	}

	fmt.Printf("%s", bytes)

	return nil
}
