package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var publishControlsCmd = &cobra.Command{
	Use:   "controls <path>",
	Short: "Publish controls artifacts to the website repository",
	Args:  cobra.ExactArgs(1),
	RunE:  runPublishControls,
}

func init() {
	publishControlsCmd.Flags().String("tag", "", "Release tag (required)")
	publishControlsCmd.Flags().String("website-repo", "", "Website repository in owner/repo format (required)")
	publishControlsCmd.Flags().String("token", "", "GitHub token with write access to the website repo (falls back to $GITHUB_TOKEN)")
	publishControlsCmd.Flags().StringP("output-dir", "o", "artifacts", "Directory containing generated artifacts")
	publishControlsCmd.MarkFlagRequired("tag")
	publishControlsCmd.MarkFlagRequired("website-repo")
}

func doPublishControls(catalogPath, tag, websiteRepo, token, outputDir string) error {
	base := filepath.Join(outputDir, catalogPath)

	mdContent, err := os.ReadFile(filepath.Join(base, "controls.md"))
	if err != nil {
		return fmt.Errorf("reading controls.md: %w", err)
	}
	yamlContent, err := os.ReadFile(filepath.Join(base, "controls.yaml"))
	if err != nil {
		return fmt.Errorf("reading controls.yaml: %w", err)
	}

	return publishArtifact(token, websiteRepo, catalogPath, tag, "controls", mdContent, yamlContent)
}

func runPublishControls(cmd *cobra.Command, args []string) error {
	tag, _ := cmd.Flags().GetString("tag")
	websiteRepo, _ := cmd.Flags().GetString("website-repo")
	outputDir, _ := cmd.Flags().GetString("output-dir")
	token, err := resolveToken(cmd)
	if err != nil {
		return err
	}
	return doPublishControls(args[0], tag, websiteRepo, token, outputDir)
}
