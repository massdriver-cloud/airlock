package cmd

import (
	"fmt"
	"os"

	"github.com/massdriver-cloud/airlock/docs/helpdocs"
	"github.com/massdriver-cloud/airlock/pkg/bicep"
	"github.com/spf13/cobra"
)

func NewCmdBicep() *cobra.Command {
	bicepCmd := &cobra.Command{
		Use:   "bicep",
		Short: "Bicep translations",
		Long:  helpdocs.MustRender("bicep"),
	}

	// Input
	bicepInputCmd := &cobra.Command{
		Use:   `input`,
		Short: "Ingest a Bicep template file and generate a JSON Schema from the params",
		Args:  cobra.ExactArgs(1),
		Long:  helpdocs.MustRender("bicep/input"),
		RunE:  runBicepInput,
	}

	// Output
	bicepOutputCmd := &cobra.Command{
		Use:   "output",
		Short: "Output a Bicep params specification from a JSON Schema document",
		Args:  cobra.ExactArgs(1),
		Long:  helpdocs.MustRender("bicep/output"),
		RunE:  runBicepOutput,
	}

	bicepCmd.AddCommand(bicepInputCmd)
	bicepCmd.AddCommand(bicepOutputCmd)

	return bicepCmd
}

func runBicepInput(cmd *cobra.Command, args []string) error {
	result := bicep.BicepToSchema(args[0])

	fmt.Print(result.PrettyDiags())
	fmt.Print(result.PrettySchema())

	return nil
}

func runBicepOutput(cmd *cobra.Command, args []string) error {
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

	bytes, err := bicep.SchemaToBicep(in)
	if err != nil {
		return err
	}

	fmt.Printf("%s", bytes)
	return nil
}
