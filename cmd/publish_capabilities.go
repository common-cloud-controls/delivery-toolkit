package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var publishCapabilitiesCmd = &cobra.Command{
	Use:   "capabilities <path>",
	Short: "Publish capabilities artifacts to the website repository",
	Args:  cobra.ExactArgs(1),
	RunE:  runPublishCapabilities,
}

func init() {
	publishCapabilitiesCmd.Flags().String("tag", "", "Release tag (required)")
	publishCapabilitiesCmd.Flags().String("website-repo", "", "Website repository in owner/repo format (required)")
	publishCapabilitiesCmd.Flags().String("token", "", "GitHub token with write access to the website repo (falls back to $GITHUB_TOKEN)")
	publishCapabilitiesCmd.Flags().StringP("output-dir", "o", "artifacts", "Directory containing generated artifacts")
	publishCapabilitiesCmd.MarkFlagRequired("tag")
	publishCapabilitiesCmd.MarkFlagRequired("website-repo")
}

func doPublishCapabilities(catalogPath, tag, websiteRepo, token, outputDir string) error {
	base := filepath.Join(outputDir, catalogPath)

	mdContent, err := os.ReadFile(filepath.Join(base, "capabilities.md"))
	if err != nil {
		return fmt.Errorf("reading capabilities.md: %w", err)
	}
	yamlContent, err := os.ReadFile(filepath.Join(base, "capabilities.yaml"))
	if err != nil {
		return fmt.Errorf("reading capabilities.yaml: %w", err)
	}

	return publishArtifact(token, websiteRepo, catalogPath, tag, "capabilities", mdContent, yamlContent)
}

func runPublishCapabilities(cmd *cobra.Command, args []string) error {
	tag, _ := cmd.Flags().GetString("tag")
	websiteRepo, _ := cmd.Flags().GetString("website-repo")
	outputDir, _ := cmd.Flags().GetString("output-dir")
	token, err := resolveToken(cmd)
	if err != nil {
		return err
	}
	return doPublishCapabilities(args[0], tag, websiteRepo, token, outputDir)
}
