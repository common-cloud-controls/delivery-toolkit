package cmd

import "github.com/spf13/cobra"

var releaseControlsCmd = &cobra.Command{
	Use:   "controls <path> <title>",
	Short: "Release YAML and Markdown from a controls catalog",
	Long:  "Reads a controls.yaml, injects CCC metadata, and writes controls.yaml and controls.md to <output-dir>/<path>/. Identical to `ccc generate controls` but intended as a named CI step that must pass before `ccc publish` runs.",
	Args:  cobra.ExactArgs(2),
	RunE:  runGenerateControls,
}

func init() {
	releaseControlsCmd.Flags().String("controls-dir", "", "Root of the control-catalogs repo (omit to fetch from GitHub)")
	releaseControlsCmd.Flags().StringP("output-dir", "o", "artifacts", "Directory to write generated files into")
	releaseControlsCmd.Flags().String("tag", "dev", "Release tag to embed in artifact metadata (e.g. v2026.04-rc)")
}
