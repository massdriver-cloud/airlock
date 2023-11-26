package cmd

import (
	"errors"
	"fmt"

	"github.com/massdriver-cloud/airlock/docs/helpdocs"
	"github.com/massdriver-cloud/airlock/pkg/validate"
	"github.com/spf13/cobra"
)

func NewCmdValidate() *cobra.Command {
	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a document matches a JSON Schema",
		RunE:  runValidate,
		Long:  helpdocs.MustRender("validate"),
	}
	validateCmd.Flags().StringP("document", "d", "document.json", "Path to document")
	validateCmd.Flags().StringP("schema", "s", "./schema.json", "Path to JSON Schema")

	return validateCmd
}

func runValidate(cmd *cobra.Command, args []string) error {
	schema, _ := cmd.Flags().GetString("schema")
	document, _ := cmd.Flags().GetString("document")

	result, err := validate.Validate(schema, document)
	if err != nil {
		return err
	}

	if result.Valid() {
		fmt.Println("The document is valid!")
	} else {
		errMsg := fmt.Sprintf("The document failed validation:\n\tDocument: %s\n\tSchema: %s\nErrors:\n", document, schema)
		for _, violation := range result.Errors() {
			errMsg += fmt.Sprintf("\t- %v\n", violation)
		}

		return errors.New(errMsg)
	}
	return nil
}
