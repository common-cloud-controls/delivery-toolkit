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

// githubGet performs an authenticated GET and decodes the JSON response into dst.
func githubGet(token, url string, dst interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}

// githubPost performs an authenticated POST and decodes the JSON response into dst.
func githubPost(token, url string, body, dst interface{}) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
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
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("POST %s: %s", url, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}

// githubPatch performs an authenticated PATCH request.
func githubPatch(token, url string, body interface{}) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PATCH", url, bytes.NewReader(b))
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
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("PATCH %s: %s", url, resp.Status)
	}
	return nil
}

// githubCreateBlob uploads content as a blob and returns its SHA.
func githubCreateBlob(token, repo string, content []byte) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/git/blobs", repo)
	var result struct {
		SHA string `json:"sha"`
	}
	err := githubPost(token, url, map[string]string{
		"content":  base64.StdEncoding.EncodeToString(content),
		"encoding": "base64",
	}, &result)
	return result.SHA, err
}

// githubCommitFiles creates a single commit on the default branch containing all provided files.
func githubCommitFiles(token, repo, message string, files map[string][]byte) error {
	base := fmt.Sprintf("https://api.github.com/repos/%s/git", repo)

	// Get current HEAD commit SHA.
	var ref struct {
		Object struct{ SHA string `json:"sha"` } `json:"object"`
	}
	if err := githubGet(token, base+"/refs/heads/main", &ref); err != nil {
		return fmt.Errorf("getting ref: %w", err)
	}

	// Get the tree SHA from that commit.
	var commit struct {
		Tree struct{ SHA string `json:"sha"` } `json:"tree"`
	}
	if err := githubGet(token, base+"/commits/"+ref.Object.SHA, &commit); err != nil {
		return fmt.Errorf("getting commit: %w", err)
	}

	// Create a blob for each file.
	type treeEntry struct {
		Path string `json:"path"`
		Mode string `json:"mode"`
		Type string `json:"type"`
		SHA  string `json:"sha"`
	}
	entries := make([]treeEntry, 0, len(files))
	for path, content := range files {
		blobSHA, err := githubCreateBlob(token, repo, content)
		if err != nil {
			return fmt.Errorf("creating blob for %s: %w", path, err)
		}
		entries = append(entries, treeEntry{Path: path, Mode: "100644", Type: "blob", SHA: blobSHA})
	}

	// Create a new tree on top of the current one.
	var tree struct {
		SHA string `json:"sha"`
	}
	if err := githubPost(token, base+"/trees", map[string]interface{}{
		"base_tree": commit.Tree.SHA,
		"tree":      entries,
	}, &tree); err != nil {
		return fmt.Errorf("creating tree: %w", err)
	}

	// Create the commit.
	var newCommit struct {
		SHA string `json:"sha"`
	}
	if err := githubPost(token, base+"/commits", map[string]interface{}{
		"message": message,
		"tree":    tree.SHA,
		"parents": []string{ref.Object.SHA},
	}, &newCommit); err != nil {
		return fmt.Errorf("creating commit: %w", err)
	}

	// Advance the branch ref.
	if err := githubPatch(token, base+"/refs/heads/main", map[string]string{
		"sha": newCommit.SHA,
	}); err != nil {
		return fmt.Errorf("updating ref: %w", err)
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
// to the website repository in a single commit.
func publishArtifact(token, websiteRepo, catalogPath, tag, artifactType string, mdContent, yamlContent []byte) error {
	mdDest := fmt.Sprintf("src/content/catalogs/%s/%s/%s.md", catalogPath, artifactType, tag)
	yamlDest := fmt.Sprintf("public/data/catalogs/%s/%s/%s.yaml", catalogPath, artifactType, tag)
	commitMsg := fmt.Sprintf("release: %s %s@%s", catalogPath, artifactType, tag)

	title := extractTitle(string(mdContent))
	pagePath := fmt.Sprintf("/catalogs/%s/%s/%s", catalogPath, artifactType, tag)
	mdWithFM := withFrontmatter(string(mdContent), title, pagePath)

	if err := githubCommitFiles(token, websiteRepo, commitMsg, map[string][]byte{
		mdDest:   mdWithFM,
		yamlDest: yamlContent,
	}); err != nil {
		return err
	}
	fmt.Printf("Published %s and %s\n", mdDest, yamlDest)
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
