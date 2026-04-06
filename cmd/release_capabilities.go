package cmd

import "github.com/spf13/cobra"

var releaseCapabilitiesCmd = &cobra.Command{
	Use:   "capabilities <path> <title>",
	Short: "Release YAML and Markdown from a capabilities catalog",
	Long:  "Identical to `ccc generate capabilities` — exists as a distinct command so CI pipelines have a named failure point before publish.",
	Args:  cobra.ExactArgs(2),
	RunE:  runGenerateCapabilities,
}

func init() {
	releaseCapabilitiesCmd.Flags().String("capabilities-dir", "", "Root of the capability-catalogs repo (omit to fetch from GitHub)")
	releaseCapabilitiesCmd.Flags().StringP("output-dir", "o", "artifacts", "Directory to write generated files into")
}
