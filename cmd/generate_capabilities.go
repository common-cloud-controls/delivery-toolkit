package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	gemara "github.com/gemaraproj/go-gemara"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var trailingDigits = regexp.MustCompile(`\d+$`)

// inferCatalogID derives the catalog ID from capability entry IDs by stripping
// the trailing numeric suffix. e.g. "CCC.ObjStor.CP01" → "CCC.ObjStor.CP"
func inferCatalogID(capabilities []gemara.Capability) string {
	if len(capabilities) == 0 {
		return "CCC"
	}
	return trailingDigits.ReplaceAllString(capabilities[0].Id, "")
}

const githubRawBase = "https://raw.githubusercontent.com/common-cloud-controls/capability-catalogs/refs/heads/main"
const githubRawCoreBase = "https://raw.githubusercontent.com/common-cloud-controls/core-catalog/refs/heads/main"

const corePath = "core/ccc"

// resolveGitHubURL returns the full GitHub raw URL for a catalog file.
// The core catalog (path "core/ccc") lives in a separate repo; all others
// use the provided repoBase.
func resolveGitHubURL(repoBase, catalogPath, filename string) string {
	if catalogPath == corePath {
		return githubRawCoreBase + "/ccc/" + filename
	}
	return repoBase + "/" + catalogPath + "/" + filename
}

// resolveLocalPath returns the filesystem path to a catalog file.
// For "core/ccc", the core-catalog repo stores files under ccc/ (not core/ccc/),
// so the caller should pass the root of that repo as dir.
func resolveLocalPath(dir, catalogPath, filename string) string {
	if catalogPath == corePath {
		return filepath.Join(dir, "ccc", filename)
	}
	return filepath.Join(dir, catalogPath, filename)
}

var generateCapabilitiesCmd = &cobra.Command{
	Use:   "capabilities <path> <title>",
	Short: "Generate YAML and Markdown from a capabilities catalog",
	Long: `Reads a capabilities.yaml at <capabilities-dir>/<path>/capabilities.yaml,
injects metadata, and writes capabilities.yaml and capabilities.md to <output-dir>/<path>/.

The title is wrapped to form: "CCC <title> Capabilities"

If --capabilities-dir is not provided, the catalog is fetched from GitHub.
For most paths: ` + githubRawBase + `/<path>/capabilities.yaml
For core/ccc:   ` + githubRawCoreBase + `/ccc/capabilities.yaml`,
	Args: cobra.ExactArgs(2),
	RunE: runGenerateCapabilities,
}

func init() {
	generateCapabilitiesCmd.Flags().String("capabilities-dir", "", "Root of the capability-catalogs repo (omit to fetch from GitHub)")
	generateCapabilitiesCmd.Flags().StringP("output-dir", "o", "artifacts", "Directory to write generated files into")
}

func runGenerateCapabilities(cmd *cobra.Command, args []string) error {
	catalogPath := args[0]
	catalogTitle := "CCC " + args[1] + " Capabilities"
	capabilitiesDir, _ := cmd.Flags().GetString("capabilities-dir")
	outputDir, _ := cmd.Flags().GetString("output-dir")

	// Load capabilities.yaml — from disk or GitHub
	var data []byte
	if capabilitiesDir != "" {
		absInput, err := filepath.Abs(resolveLocalPath(capabilitiesDir, catalogPath, "capabilities.yaml"))
		if err != nil {
			return fmt.Errorf("resolving input path: %w", err)
		}
		data, err = os.ReadFile(absInput)
		if err != nil {
			return fmt.Errorf("reading %s: %w", absInput, err)
		}
	} else {
		url := resolveGitHubURL(githubRawBase, catalogPath, "capabilities.yaml")
		var err error
		data, err = fetchURL(url)
		if err != nil {
			return fmt.Errorf("fetching %s: %w", url, err)
		}
	}

	var catalog gemara.CapabilityCatalog
	if err := yaml.Unmarshal(data, &catalog); err != nil {
		return fmt.Errorf("parsing capabilities.yaml: %w", err)
	}

	// Inject hardcoded metadata
	// TODO: replace ControlCatalogArtifact with a CapabilityCatalogArtifact once added to go-gemara
	catalog.Title = catalogTitle
	catalog.Metadata = gemara.Metadata{
		Id:            inferCatalogID(catalog.Capabilities),
		Type:          gemara.ControlCatalogArtifact,
		GemaraVersion: "v0",
		Description:   "Capabilities for " + args[1] + " technologies, as defined by the FINOS Common Cloud Controls project.",
		Author: gemara.Actor{
			Id:   "FINOS-CCC",
			Name: "FINOS Common Cloud Controls",
			Type: gemara.Human,
		},
	}

	// Prepare output directory
	outDir := filepath.Join(outputDir, catalogPath)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Write YAML
	yamlOut, err := yaml.Marshal(&catalog)
	if err != nil {
		return fmt.Errorf("marshaling YAML: %w", err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "capabilities.yaml"), yamlOut, 0644); err != nil {
		return fmt.Errorf("writing capabilities.yaml: %w", err)
	}

	// Write Markdown
	md, err := renderMarkdown(&catalog)
	if err != nil {
		return fmt.Errorf("rendering Markdown: %w", err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "capabilities.md"), []byte(md), 0644); err != nil {
		return fmt.Errorf("writing capabilities.md: %w", err)
	}

	fmt.Printf("Generated artifacts in %s\n", outDir)
	return nil
}

func fetchURL(url string) ([]byte, error) {
	resp, err := http.Get(url) //nolint:gosec // URL is constructed from a fixed base and a user-supplied path
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}
