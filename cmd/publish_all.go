package cmd

import (
	"github.com/spf13/cobra"
)

var publishAllCmd = &cobra.Command{
	Use:   "all <path>",
	Short: "Publish all catalog types (capabilities, threats, controls) to the website repository",
	Args:  cobra.ExactArgs(1),
	RunE:  runPublishAll,
}

func init() {
	publishAllCmd.Flags().String("tag", "", "Release tag (required)")
	publishAllCmd.Flags().String("website-repo", "", "Website repository in owner/repo format (required)")
	publishAllCmd.Flags().String("token", "", "GitHub token with write access to the website repo (falls back to $GITHUB_TOKEN)")
	publishAllCmd.Flags().StringP("output-dir", "o", "artifacts", "Directory containing generated artifacts")
	publishAllCmd.MarkFlagRequired("tag")
	publishAllCmd.MarkFlagRequired("website-repo")
}

func runPublishAll(cmd *cobra.Command, args []string) error {
	catalogPath := args[0]
	tag, _ := cmd.Flags().GetString("tag")
	websiteRepo, _ := cmd.Flags().GetString("website-repo")
	outputDir, _ := cmd.Flags().GetString("output-dir")
	token, err := resolveToken(cmd)
	if err != nil {
		return err
	}

	if err := doPublishCapabilities(catalogPath, tag, websiteRepo, token, outputDir); err != nil {
		return err
	}
	if err := doPublishThreats(catalogPath, tag, websiteRepo, token, outputDir); err != nil {
		return err
	}
	return doPublishControls(catalogPath, tag, websiteRepo, token, outputDir)
}
