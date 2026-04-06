package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var publishThreatsCmd = &cobra.Command{
	Use:   "threats <path>",
	Short: "Publish threats artifacts to the website repository",
	Args:  cobra.ExactArgs(1),
	RunE:  runPublishThreats,
}

func init() {
	publishThreatsCmd.Flags().String("tag", "", "Release tag (required)")
	publishThreatsCmd.Flags().String("website-repo", "", "Website repository in owner/repo format (required)")
	publishThreatsCmd.Flags().String("token", "", "GitHub token with write access to the website repo (falls back to $GITHUB_TOKEN)")
	publishThreatsCmd.Flags().StringP("output-dir", "o", "artifacts", "Directory containing generated artifacts")
	publishThreatsCmd.MarkFlagRequired("tag")
	publishThreatsCmd.MarkFlagRequired("website-repo")
}

func doPublishThreats(catalogPath, tag, websiteRepo, token, outputDir string) error {
	base := filepath.Join(outputDir, catalogPath)

	mdContent, err := os.ReadFile(filepath.Join(base, "threats.md"))
	if err != nil {
		return fmt.Errorf("reading threats.md: %w", err)
	}
	yamlContent, err := os.ReadFile(filepath.Join(base, "threats.yaml"))
	if err != nil {
		return fmt.Errorf("reading threats.yaml: %w", err)
	}

	return publishArtifact(token, websiteRepo, catalogPath, tag, "threats", mdContent, yamlContent)
}

func runPublishThreats(cmd *cobra.Command, args []string) error {
	tag, _ := cmd.Flags().GetString("tag")
	websiteRepo, _ := cmd.Flags().GetString("website-repo")
	outputDir, _ := cmd.Flags().GetString("output-dir")
	token, err := resolveToken(cmd)
	if err != nil {
		return err
	}
	return doPublishThreats(args[0], tag, websiteRepo, token, outputDir)
}
