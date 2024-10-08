package cmd

import (
	"fmt"

	"github.com/massdriver-cloud/airlock/pkg/prettylogs"
	"github.com/massdriver-cloud/airlock/pkg/version"
	"github.com/spf13/cobra"
)

func NewCmdVersion() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "Version of Airlock",
		Run:     runVersion,
	}
	return versionCmd
}

func runVersion(cmd *cobra.Command, args []string) {
	latestVersion, err := version.GetLatestVersion()
	if err != nil {
		fmt.Errorf("could not check for newer version, skipping.\nurl: %s\nerror: %w\n", version.LatestReleaseURL, err)
	}

	isOld, _ := version.CheckForNewerVersionAvailable(latestVersion)
	if isOld {
		fmt.Printf("A newer version of Airlock is available, you can download it here: %v\n", version.LatestReleaseURL)
	}
	airlockVersion := prettylogs.Green(version.AirlockVersion())
	fmt.Printf("Airlock version: %v (git SHA: %v)\n", airlockVersion, version.AirlockGitSHA())
}
