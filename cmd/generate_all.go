package cmd

import (
	"github.com/spf13/cobra"
)

var generateAllCmd = &cobra.Command{
	Use:   "all <path> <title>",
	Short: "Generate all catalog types (capabilities, threats, controls)",
	Args:  cobra.ExactArgs(2),
	RunE:  runReleaseAll, // same logic as release all
}

func init() {
	generateAllCmd.Flags().String("capabilities-dir", "", "Root of the capability-catalogs repo (omit to fetch from GitHub)")
	generateAllCmd.Flags().String("threats-dir", "", "Root of the threat-catalogs repo (omit to fetch from GitHub)")
	generateAllCmd.Flags().String("controls-dir", "", "Root of the control-catalogs repo (omit to fetch from GitHub)")
	generateAllCmd.Flags().StringP("output-dir", "o", "artifacts", "Directory to write generated files into")
}
