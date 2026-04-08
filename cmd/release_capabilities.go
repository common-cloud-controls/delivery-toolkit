package cmd

import "github.com/spf13/cobra"

var releaseCapabilitiesCmd = &cobra.Command{
	Use:   "capabilities <path> <title>",
	Short: "Release YAML and Markdown from a capabilities catalog",
	Long:  "Reads a capabilities.yaml, injects CCC metadata, and writes capabilities.yaml and capabilities.md to <output-dir>/<path>/. Identical to `ccc generate capabilities` but intended as a named CI step that must pass before `ccc publish` runs.",
	Args:  cobra.ExactArgs(2),
	RunE:  runGenerateCapabilities,
}

func init() {
	releaseCapabilitiesCmd.Flags().String("capabilities-dir", "", "Root of the capability-catalogs repo (omit to fetch from GitHub)")
	releaseCapabilitiesCmd.Flags().StringP("output-dir", "o", "artifacts", "Directory to write generated files into")
	releaseCapabilitiesCmd.Flags().String("tag", "dev", "Release tag to embed in artifact metadata (e.g. v2026.04-rc)")
}
