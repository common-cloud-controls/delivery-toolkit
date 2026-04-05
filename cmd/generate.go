package cmd

import "github.com/spf13/cobra"

// GenerateCmd is the `ccc generate` subcommand group.
var GenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate artifacts from CCC catalogs",
}

func init() {
	GenerateCmd.AddCommand(generateCapabilitiesCmd)
	GenerateCmd.AddCommand(generateThreatsCmd)
}
