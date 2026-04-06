package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	gemara "github.com/gemaraproj/go-gemara"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const githubRawControlsBase = "https://raw.githubusercontent.com/common-cloud-controls/control-catalogs/refs/heads/main"

var generateControlsCmd = &cobra.Command{
	Use:   "controls <path> <title>",
	Short: "Generate YAML and Markdown from a controls catalog",
	Long: `Reads a controls.yaml at <controls-dir>/<path>/controls.yaml,
injects metadata, and writes controls.yaml and controls.md to <output-dir>/<path>/.

The title is wrapped to form: "CCC <title> Controls"

If --controls-dir is not provided, the catalog is fetched from GitHub.
For most paths: ` + githubRawControlsBase + `/<path>/controls.yaml
For core/ccc:   ` + githubRawCoreBase + `/controls.yaml`,
	Args: cobra.ExactArgs(2),
	RunE: runGenerateControls,
}

func init() {
	generateControlsCmd.Flags().String("controls-dir", "", "Root of the control-catalogs repo (omit to fetch from GitHub)")
	generateControlsCmd.Flags().StringP("output-dir", "o", "artifacts", "Directory to write generated files into")
}

func runGenerateControls(cmd *cobra.Command, args []string) error {
	controlsDir, _ := cmd.Flags().GetString("controls-dir")
	outputDir, _ := cmd.Flags().GetString("output-dir")
	return doGenerateControls(args[0], "CCC "+args[1]+" Controls", args[1], controlsDir, outputDir)
}

func doGenerateControls(catalogPath, catalogTitle, serviceTitle, controlsDir, outputDir string) error {
	// Load controls.yaml — from disk or GitHub
	var data []byte
	if controlsDir != "" {
		absInput, err := filepath.Abs(resolveLocalPath(controlsDir, catalogPath, "controls.yaml"))
		if err != nil {
			return fmt.Errorf("resolving input path: %w", err)
		}
		data, err = os.ReadFile(absInput)
		if err != nil {
			return fmt.Errorf("reading %s: %w", absInput, err)
		}
	} else {
		url := resolveGitHubURL(githubRawControlsBase, catalogPath, "controls.yaml")
		var err error
		data, err = fetchURL(url)
		if err != nil {
			return fmt.Errorf("fetching %s: %w", url, err)
		}
	}

	var catalog gemara.ControlCatalog
	if err := yaml.Unmarshal(data, &catalog); err != nil {
		return fmt.Errorf("parsing controls.yaml: %w", err)
	}

	// Inject hardcoded metadata
	catalog.Title = catalogTitle
	catalog.Metadata = gemara.Metadata{
		Id:            inferControlCatalogID(catalog.Controls),
		Type:          gemara.ControlCatalogArtifact,
		GemaraVersion: "v0",
		Description:   "Controls for " + serviceTitle + " technologies, as defined by the FINOS Common Cloud Controls project.",
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
	if err := os.WriteFile(filepath.Join(outDir, "controls.yaml"), yamlOut, 0644); err != nil {
		return fmt.Errorf("writing controls.yaml: %w", err)
	}

	// Write Markdown
	md, err := renderControlsMarkdown(&catalog)
	if err != nil {
		return fmt.Errorf("rendering Markdown: %w", err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "controls.md"), []byte(md), 0644); err != nil {
		return fmt.Errorf("writing controls.md: %w", err)
	}

	fmt.Printf("Generated artifacts in %s\n", outDir)
	return nil
}

// inferControlCatalogID derives the catalog ID from control entry IDs by stripping
// the trailing numeric suffix. e.g. "CCC.ObjStor.CN01" → "CCC.ObjStor.CN"
func inferControlCatalogID(controls []gemara.Control) string {
	if len(controls) == 0 {
		return "CCC"
	}
	return trailingDigits.ReplaceAllString(controls[0].Id, "")
}
