package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/massdriver-cloud/airlock/docs/helpdocs"
	"github.com/massdriver-cloud/airlock/pkg/opentofu"
	"github.com/spf13/cobra"
)

func NewCmdOpenTofu() *cobra.Command {
	opentofuCmd := &cobra.Command{
		Use:     "opentofu",
		Short:   "OpenTofu (HCL) translations",
		Aliases: []string{"tofu", "tf", "terraform"},
		Long:    helpdocs.MustRender("opentofu"),
	}

	// Input
	opentofuInputCmd := &cobra.Command{
		Use:   `input`,
		Short: "Ingest a OpenTofu module and generate a JSON Schema from the variables",
		Args:  cobra.ExactArgs(1),
		Long:  helpdocs.MustRender("opentofu/input"),
		RunE:  runOpenTofuInput,
	}

	// oputput
	opentofuOutputCmd := &cobra.Command{
		Use:   `output`,
		Short: "Output a OpenTofu variables specification from a JSON schemea document",
		Args:  cobra.ExactArgs(1),
		Long:  helpdocs.MustRender("opentofu/output"),
		RunE:  runOpenTofuOutput,
	}

	opentofuCmd.AddCommand(opentofuInputCmd)
	opentofuCmd.AddCommand(opentofuOutputCmd)

	return opentofuCmd
}

func runOpenTofuInput(cmd *cobra.Command, args []string) error {
	schema, err := opentofu.TfToSchema(args[0])
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

func runOpenTofuOutput(cmd *cobra.Command, args []string) error {
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

	bytes, err := opentofu.SchemaToTf(in)
	if err != nil {
		return err
	}

	fmt.Printf("%s", bytes)

	return nil
}
