package cmd

import "github.com/spf13/cobra"

var releaseThreatsCmd = &cobra.Command{
	Use:   "threats <path> <title>",
	Short: "Release YAML and Markdown from a threats catalog",
	Long:  "Reads a threats.yaml, injects CCC metadata, and writes threats.yaml and threats.md to <output-dir>/<path>/. Identical to `ccc generate threats` but intended as a named CI step that must pass before `ccc publish` runs.",
	Args:  cobra.ExactArgs(2),
	RunE:  runGenerateThreats,
}

func init() {
	releaseThreatsCmd.Flags().String("threats-dir", "", "Root of the threat-catalogs repo (omit to fetch from GitHub)")
	releaseThreatsCmd.Flags().StringP("output-dir", "o", "artifacts", "Directory to write generated files into")
	releaseThreatsCmd.Flags().String("tag", "dev", "Release tag to embed in artifact metadata (e.g. v2026.04-rc)")
}
