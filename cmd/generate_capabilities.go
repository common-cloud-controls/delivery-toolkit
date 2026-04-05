package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	gemara "github.com/gemaraproj/go-gemara"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const githubRawBase = "https://raw.githubusercontent.com/common-cloud-controls/capability-catalogs/refs/heads/main"

var generateCapabilitiesCmd = &cobra.Command{
	Use:   "capabilities <path>",
	Short: "Generate YAML and Markdown from a capabilities catalog",
	Long: `Reads a capabilities.yaml at <capabilities-dir>/<path>/capabilities.yaml,
injects metadata, and writes capabilities.yaml and capabilities.md to <output-dir>/<path>/.

If --capabilities-dir is not provided, the catalog is fetched from:
  ` + githubRawBase + `/<path>/capabilities.yaml`,
	Args: cobra.ExactArgs(1),
	RunE: runGenerateCapabilities,
}

func init() {
	generateCapabilitiesCmd.Flags().String("capabilities-dir", "", "Root of the capability-catalogs repo (omit to fetch from GitHub)")
	generateCapabilitiesCmd.Flags().StringP("output-dir", "o", "artifacts", "Directory to write generated files into")
}

func runGenerateCapabilities(cmd *cobra.Command, args []string) error {
	catalogPath := args[0]
	capabilitiesDir, _ := cmd.Flags().GetString("capabilities-dir")
	outputDir, _ := cmd.Flags().GetString("output-dir")

	// Load capabilities.yaml — from disk or GitHub
	var data []byte
	var err error
	if capabilitiesDir != "" {
		inputFile := filepath.Join(capabilitiesDir, catalogPath, "capabilities.yaml")
		absInput, err := filepath.Abs(inputFile)
		if err != nil {
			return fmt.Errorf("resolving input path: %w", err)
		}
		data, err = os.ReadFile(absInput)
		if err != nil {
			return fmt.Errorf("reading %s: %w", absInput, err)
		}
	} else {
		url := githubRawBase + "/" + catalogPath + "/capabilities.yaml"
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
	catalog.Title = catalogPath
	catalog.Metadata = gemara.Metadata{
		Id:            "CCC",
		Type:          gemara.ControlCatalogArtifact,
		GemaraVersion: "v0",
		Description:   "Capability catalog for " + catalogPath,
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
