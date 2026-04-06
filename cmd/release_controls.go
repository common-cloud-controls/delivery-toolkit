package cmd

import "github.com/spf13/cobra"

var releaseControlsCmd = &cobra.Command{
	Use:   "controls <path> <title>",
	Short: "Release YAML and Markdown from a controls catalog",
	Long:  "Identical to `ccc generate controls` — exists as a distinct command so CI pipelines have a named failure point before publish.",
	Args:  cobra.ExactArgs(2),
	RunE:  runGenerateControls,
}

func init() {
	releaseControlsCmd.Flags().String("controls-dir", "", "Root of the control-catalogs repo (omit to fetch from GitHub)")
	releaseControlsCmd.Flags().StringP("output-dir", "o", "artifacts", "Directory to write generated files into")
}
