package cmd

import (
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
		Short: "Ingest an OpenTofu module and generate a JSON Schema from the variables",
		Args:  cobra.ExactArgs(1),
		Long:  helpdocs.MustRender("opentofu/input"),
		RunE:  runOpenTofuInput,
	}

	// Output
	opentofuOutputCmd := &cobra.Command{
		Use:   `output`,
		Short: "Output an OpenTofu variables specification from a JSON Schema document",
		Args:  cobra.ExactArgs(1),
		Long:  helpdocs.MustRender("opentofu/output"),
		RunE:  runOpenTofuOutput,
	}

	opentofuCmd.AddCommand(opentofuInputCmd)
	opentofuCmd.AddCommand(opentofuOutputCmd)

	return opentofuCmd
}

func runOpenTofuInput(cmd *cobra.Command, args []string) error {
	result := opentofu.TofuToSchema(args[0])

	fmt.Print(result.PrettyDiags())
	fmt.Print(result.PrettySchema())

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

	bytes, err := opentofu.SchemaToTofu(in)
	if err != nil {
		return err
	}

	fmt.Printf("%s", bytes)

	return nil
}
