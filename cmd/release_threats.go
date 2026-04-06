package cmd

import "github.com/spf13/cobra"

var releaseThreatsCmd = &cobra.Command{
	Use:   "threats <path> <title>",
	Short: "Release YAML and Markdown from a threats catalog",
	Long:  "Identical to `ccc generate threats` — exists as a distinct command so CI pipelines have a named failure point before publish.",
	Args:  cobra.ExactArgs(2),
	RunE:  runGenerateThreats,
}

func init() {
	releaseThreatsCmd.Flags().String("threats-dir", "", "Root of the threat-catalogs repo (omit to fetch from GitHub)")
	releaseThreatsCmd.Flags().StringP("output-dir", "o", "artifacts", "Directory to write generated files into")
}
