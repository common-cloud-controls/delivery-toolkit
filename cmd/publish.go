package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// PublishCmd is the `ccc publish` subcommand group.
var PublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish released catalog artifacts to the CCC website repository",
}

func init() {
	PublishCmd.AddCommand(publishCapabilitiesCmd)
	PublishCmd.AddCommand(publishThreatsCmd)
	PublishCmd.AddCommand(publishControlsCmd)
	PublishCmd.AddCommand(publishAllCmd)
}

// githubFile is the subset of the GitHub Contents API GET response we care about.
type githubFile struct {
	SHA string `json:"sha"`
}

// githubGetSHA returns the current blob SHA of a file in a GitHub repo, or ""
// if the file does not exist. SHA is required by the API when updating an existing file.
func githubGetSHA(token, repo, path string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s", repo, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", nil
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET %s: %s", url, resp.Status)
	}

	var f githubFile
	if err := json.NewDecoder(resp.Body).Decode(&f); err != nil {
		return "", err
	}
	return f.SHA, nil
}

// githubPutFile creates or updates a file in a GitHub repo via the Contents API.
// Pass sha="" when creating a new file; pass the existing SHA when updating.
func githubPutFile(token, repo, filePath, commitMsg string, content []byte, sha string) error {
	type putBody struct {
		Message string `json:"message"`
		Content string `json:"content"`
		SHA     string `json:"sha,omitempty"`
	}

	body := putBody{
		Message: commitMsg,
		Content: base64.StdEncoding.EncodeToString(content),
		SHA:     sha,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s", repo, filePath)
	req, err := http.NewRequest("PUT", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("PUT %s: %s", url, resp.Status)
	}
	return nil
}

// extractTitle reads the first `# ` heading from a markdown string.
func extractTitle(md string) string {
	for _, line := range strings.Split(md, "\n") {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return "CCC Catalog"
}

// withFrontmatter prepends YAML frontmatter to a markdown string so the website
// can route and title the page correctly.
func withFrontmatter(md, title, pagePath string) []byte {
	fm := fmt.Sprintf("---\ntitle: %q\npath: %q\n---\n\n", title, pagePath)
	return []byte(fm + md)
}

// publishArtifact commits a single catalog type's markdown and YAML artifacts
// to the website repository under the versioned paths.
func publishArtifact(token, websiteRepo, catalogPath, tag, artifactType string, mdContent, yamlContent []byte) error {
	mdDest := fmt.Sprintf("src/content/catalogs/%s/%s/%s.md", catalogPath, artifactType, tag)
	yamlDest := fmt.Sprintf("public/data/catalogs/%s/%s/%s.yaml", catalogPath, artifactType, tag)
	commitMsg := fmt.Sprintf("release: %s %s@%s", catalogPath, artifactType, tag)

	title := extractTitle(string(mdContent))
	pagePath := fmt.Sprintf("/catalogs/%s/%s/%s", catalogPath, artifactType, tag)
	mdWithFM := withFrontmatter(string(mdContent), title, pagePath)

	// Publish markdown
	mdSHA, err := githubGetSHA(token, websiteRepo, mdDest)
	if err != nil {
		return fmt.Errorf("checking %s: %w", mdDest, err)
	}
	if err := githubPutFile(token, websiteRepo, mdDest, commitMsg, mdWithFM, mdSHA); err != nil {
		return fmt.Errorf("publishing %s: %w", mdDest, err)
	}
	fmt.Printf("Published %s\n", mdDest)

	// Publish YAML
	yamlSHA, err := githubGetSHA(token, websiteRepo, yamlDest)
	if err != nil {
		return fmt.Errorf("checking %s: %w", yamlDest, err)
	}
	if err := githubPutFile(token, websiteRepo, yamlDest, commitMsg, yamlContent, yamlSHA); err != nil {
		return fmt.Errorf("publishing %s: %w", yamlDest, err)
	}
	fmt.Printf("Published %s\n", yamlDest)

	return nil
}

// resolveToken returns the token flag value, falling back to $GITHUB_TOKEN.
func resolveToken(cmd *cobra.Command) (string, error) {
	token, _ := cmd.Flags().GetString("token")
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token == "" {
		return "", fmt.Errorf("GitHub token required: set --token or GITHUB_TOKEN")
	}
	return token, nil
}
