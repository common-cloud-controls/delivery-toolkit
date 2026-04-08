package cmd

import (
	"fmt"
	"os"
	"path/filepath"

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

	base := filepath.Join(outputDir, catalogPath)
	files := make(map[string][]byte)

	types := []string{"capabilities", "threats", "controls"}
	for _, t := range types {
		mdContent, err := os.ReadFile(filepath.Join(base, t+".md"))
		if err != nil {
			return fmt.Errorf("reading %s.md: %w", t, err)
		}
		yamlContent, err := os.ReadFile(filepath.Join(base, t+".yaml"))
		if err != nil {
			return fmt.Errorf("reading %s.yaml: %w", t, err)
		}

		title := extractTitle(string(mdContent))
		pagePath := fmt.Sprintf("/catalogs/%s/%s/%s", catalogPath, t, tag)
		mdDest := fmt.Sprintf("src/content/catalogs/%s/%s/%s.md", catalogPath, t, tag)
		yamlDest := fmt.Sprintf("public/data/catalogs/%s/%s/%s.yaml", catalogPath, t, tag)

		files[mdDest] = withFrontmatter(string(mdContent), title, pagePath)
		files[yamlDest] = yamlContent
	}

	commitMsg := fmt.Sprintf("release: %s all@%s", catalogPath, tag)
	if err := githubCommitFiles(token, websiteRepo, commitMsg, files); err != nil {
		return err
	}

	for path := range files {
		fmt.Printf("Published %s\n", path)
	}
	return nil
}
