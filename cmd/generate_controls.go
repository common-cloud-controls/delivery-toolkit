package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	gemara "github.com/gemaraproj/go-gemara"
	"github.com/gemaraproj/go-gemara/gemaraconv"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const githubRawControlsBase = "https://raw.githubusercontent.com/common-cloud-controls/control-catalogs/refs/heads/main"

// knownControlGroups defines control family groups that should be injected into
// a catalog's Groups when referenced by one or more controls.
var knownControlGroups = map[string]gemara.Group{
	"CCC.Core.Data": {
		Id:    "CCC.Core.Data",
		Title: "Data",
		Description: "The Data control family ensures the confidentiality, integrity,\n" +
			"availability, and sovereignty of data across its lifecycle.\n" +
			"These controls govern how data is transmitted, stored,\n" +
			"replicated, and protected from unauthorized access, tampering,\n" +
			"or exposure beyond defined trust perimeters.\n",
	},
	"CCC.Core.IAM": {
		Id:    "CCC.Core.IAM",
		Title: "Identity and Access Management",
		Description: "The Identity and Access Management control family ensures\n" +
			"that only trusted and authenticated entities can access\n" +
			"resources. These controls establish strong authentication,\n" +
			"enforce multi-factor verification, and restrict access to\n" +
			"approved sources to prevent unauthorized use or data exfiltration.\n",
	},
	"CCC.Core.LM": {
		Id:    "CCC.Core.LM",
		Title: "Logging & Monitoring",
		Description: "The Logging & Monitoring control family ensures that access,\n" +
			"changes, and security-relevant events are captured, monitored,\n" +
			"and alerted on in order to provide visibility, support\n" +
			"incident response, and meet compliance requirements.\n",
	},
}

// injectControlGroups adds known group definitions to the catalog's Groups
// for any group IDs referenced by controls that aren't already present.
func injectControlGroups(catalog *gemara.ControlCatalog) {
	existing := map[string]bool{}
	for _, g := range catalog.Groups {
		existing[g.Id] = true
	}
	for _, ctrl := range catalog.Controls {
		if ctrl.Group == "" || existing[ctrl.Group] {
			continue
		}
		if g, ok := knownControlGroups[ctrl.Group]; ok {
			catalog.Groups = append(catalog.Groups, g)
			existing[ctrl.Group] = true
		}
	}
}

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
	generateControlsCmd.Flags().String("tag", "dev", "Release tag to embed in artifact metadata (e.g. v2026.04-rc)")
}

func runGenerateControls(cmd *cobra.Command, args []string) error {
	controlsDir, _ := cmd.Flags().GetString("controls-dir")
	outputDir, _ := cmd.Flags().GetString("output-dir")
	tag, _ := cmd.Flags().GetString("tag")
	return doGenerateControls(args[0], "CCC "+args[1]+" Controls", args[1], controlsDir, outputDir, tag)
}

func doGenerateControls(catalogPath, catalogTitle, serviceTitle, controlsDir, outputDir, tag string) error {
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
	catalogID, err := inferControlCatalogID(catalog.Controls)
	if err != nil {
		return err
	}

	catalog.Title = catalogTitle
	catalog.Metadata = gemara.Metadata{
		Id:                catalogID,
		Type:              gemara.ControlCatalogArtifact,
		GemaraVersion:     gemara.SchemaVersion,
		Version:           tag,
		Description:       "Controls for " + serviceTitle + " technologies, as defined by the FINOS Common Cloud Controls project.",
		MappingReferences: mappingRefsFromImports(catalog.Imports, tag),
		Author: gemara.Actor{
			Id:   "FINOS-CCC",
			Name: "FINOS Common Cloud Controls",
			Type: gemara.Human,
		},
	}

	// Inject group definitions for any referenced control families
	injectControlGroups(&catalog)

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

	// Write Markdown using the SDK's built-in renderer
	md, err := gemaraconv.ControlCatalog(&catalog).ToMarkdown(context.Background())
	if err != nil {
		return fmt.Errorf("rendering Markdown: %w", err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "controls.md"), md, 0644); err != nil {
		return fmt.Errorf("writing controls.md: %w", err)
	}

	fmt.Printf("Generated artifacts in %s\n", outDir)
	return nil
}

// inferControlCatalogID derives the catalog ID from control entry IDs by stripping
// the trailing numeric suffix. e.g. "CCC.ObjStor.CN01" → "CCC.ObjStor.CN"
// Core catalogs are mapped to their long canonical IDs.
func inferControlCatalogID(controls []gemara.Control) (string, error) {
	if len(controls) == 0 {
		return "", fmt.Errorf("cannot infer catalog ID: controls list is empty")
	}
	short := trailingDigits.ReplaceAllString(controls[0].Id, "")
	return short, nil
}
