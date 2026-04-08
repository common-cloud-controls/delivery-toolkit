package cmd

import (
	"github.com/spf13/cobra"
)

var releaseAllCmd = &cobra.Command{
	Use:   "all <path> <title>",
	Short: "Release all catalog types (capabilities, threats, controls)",
	Args:  cobra.ExactArgs(2),
	RunE:  runReleaseAll,
}

func init() {
	releaseAllCmd.Flags().String("capabilities-dir", "", "Root of the capability-catalogs repo (omit to fetch from GitHub)")
	releaseAllCmd.Flags().String("threats-dir", "", "Root of the threat-catalogs repo (omit to fetch from GitHub)")
	releaseAllCmd.Flags().String("controls-dir", "", "Root of the control-catalogs repo (omit to fetch from GitHub)")
	releaseAllCmd.Flags().StringP("output-dir", "o", "artifacts", "Directory to write generated files into")
	releaseAllCmd.Flags().String("tag", "dev", "Release tag to embed in artifact metadata (e.g. v2026.04-rc)")
}

func runReleaseAll(cmd *cobra.Command, args []string) error {
	catalogPath := args[0]
	serviceTitle := args[1]
	capabilitiesDir, _ := cmd.Flags().GetString("capabilities-dir")
	threatsDir, _ := cmd.Flags().GetString("threats-dir")
	controlsDir, _ := cmd.Flags().GetString("controls-dir")
	outputDir, _ := cmd.Flags().GetString("output-dir")
	tag, _ := cmd.Flags().GetString("tag")

	if err := doGenerateCapabilities(catalogPath, "CCC "+serviceTitle+" Capabilities", serviceTitle, capabilitiesDir, outputDir, tag); err != nil {
		return err
	}
	if err := doGenerateThreats(catalogPath, "CCC "+serviceTitle+" Threats", serviceTitle, threatsDir, outputDir, tag); err != nil {
		return err
	}
	return doGenerateControls(catalogPath, "CCC "+serviceTitle+" Controls", serviceTitle, controlsDir, outputDir, tag)
}
