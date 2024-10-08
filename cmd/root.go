/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// This has 4 spaces at the beginning to make it look nice in md. It
// turns it into a code block which preserves spaces/returns
var rootCmdHelp = `
        ___    ________  __    ____  ________ __
       /   |  /  _/ __ \/ /   / __ \/ ____/ //_/
      / /| |  / // /_/ / /   / / / / /   / ,<   
     / ___ |_/ // _, _/ /___/ /_/ / /___/ /| |  
    /_/  |_/___/_/ |_/_____/\____/\____/_/ |_|  

Translate between JSON Schema and common IaC languages
`

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:               "airlock",
	Short:             "JSON Schema tranlation tool",
	Long:              rootCmdHelp,
	DisableAutoGenTag: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.AddCommand(NewCmdBicep())
	rootCmd.AddCommand(NewCmdHelm())
	rootCmd.AddCommand(NewCmdOpenTofu())
	rootCmd.AddCommand(NewCmdValidate())
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
}
