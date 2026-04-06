package cmd

import "github.com/spf13/cobra"

// ReleaseCmd is the `ccc release` subcommand group.
var ReleaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Release CCC catalog artifacts",
	Long:  "Generate and validate catalog artifacts for release. Fails fast before any publish step.",
}

func init() {
	ReleaseCmd.AddCommand(releaseCapabilitiesCmd)
	ReleaseCmd.AddCommand(releaseThreatsCmd)
	ReleaseCmd.AddCommand(releaseControlsCmd)
	ReleaseCmd.AddCommand(releaseAllCmd)
}
