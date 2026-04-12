package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	gemara "github.com/gemaraproj/go-gemara"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const githubRawThreatsBase = "https://raw.githubusercontent.com/common-cloud-controls/threat-catalogs/refs/heads/main"

var generateThreatsCmd = &cobra.Command{
	Use:   "threats <path> <title>",
	Short: "Generate YAML and Markdown from a threats catalog",
	Long: `Reads a threats.yaml at <threats-dir>/<path>/threats.yaml,
injects metadata, and writes threats.yaml and threats.md to <output-dir>/<path>/.

The title is wrapped to form: "CCC <title> Threats"

If --threats-dir is not provided, the catalog is fetched from GitHub.
For most paths: ` + githubRawThreatsBase + `/<path>/threats.yaml
For core/ccc:   ` + githubRawCoreBase + `/threats.yaml

Note: source files must use the 'imports' key (not 'imported-threats') for
imported threats to be parsed. See the threat-catalogs migration.`,
	Args: cobra.ExactArgs(2),
	RunE: runGenerateThreats,
}

func init() {
	generateThreatsCmd.Flags().String("threats-dir", "", "Root of the threat-catalogs repo (omit to fetch from GitHub)")
	generateThreatsCmd.Flags().StringP("output-dir", "o", "artifacts", "Directory to write generated files into")
	generateThreatsCmd.Flags().String("tag", "dev", "Release tag to embed in artifact metadata (e.g. v2026.04-rc)")
}

func runGenerateThreats(cmd *cobra.Command, args []string) error {
	threatsDir, _ := cmd.Flags().GetString("threats-dir")
	outputDir, _ := cmd.Flags().GetString("output-dir")
	tag, _ := cmd.Flags().GetString("tag")
	return doGenerateThreats(args[0], "CCC "+args[1]+" Threats", args[1], threatsDir, outputDir, tag)
}

func doGenerateThreats(catalogPath, catalogTitle, serviceTitle, threatsDir, outputDir, tag string) error {
	// Load threats.yaml — from disk or GitHub
	var data []byte
	if threatsDir != "" {
		absInput, err := filepath.Abs(resolveLocalPath(threatsDir, catalogPath, "threats.yaml"))
		if err != nil {
			return fmt.Errorf("resolving input path: %w", err)
		}
		data, err = os.ReadFile(absInput)
		if err != nil {
			return fmt.Errorf("reading %s: %w", absInput, err)
		}
	} else {
		url := resolveGitHubURL(githubRawThreatsBase, catalogPath, "threats.yaml")
		var err error
		data, err = fetchURL(url)
		if err != nil {
			return fmt.Errorf("fetching %s: %w", url, err)
		}
	}

	var catalog gemara.ThreatCatalog
	if err := yaml.Unmarshal(data, &catalog); err != nil {
		return fmt.Errorf("parsing threats.yaml: %w", err)
	}

	// Inject hardcoded metadata
	catalogID, err := inferThreatCatalogID(catalog.Threats)
	if err != nil {
		return err
	}

	catalog.Title = catalogTitle
	catalog.Metadata = gemara.Metadata{
		Id:                catalogID,
		Type:              gemara.ThreatCatalogArtifact,
		GemaraVersion:     gemara.SchemaVersion,
		Version:           tag,
		Description:       "Threats for " + serviceTitle + " technologies, as defined by the FINOS Common Cloud Controls project.",
		MappingReferences: mappingRefsFromImports(catalog.Imports, tag),
		Author: gemara.Actor{
			Id:   "FINOS-CCC",
			Name: "FINOS Common Cloud Controls",
			Type: gemara.Human,
		},
	}

	// Inject group definitions for any referenced threat families
	injectThreatGroups(&catalog)

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
	if err := os.WriteFile(filepath.Join(outDir, "threats.yaml"), yamlOut, 0644); err != nil {
		return fmt.Errorf("writing threats.yaml: %w", err)
	}

	// Write Markdown
	md, err := renderThreatsMarkdown(&catalog)
	if err != nil {
		return fmt.Errorf("rendering Markdown: %w", err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "threats.md"), []byte(md), 0644); err != nil {
		return fmt.Errorf("writing threats.md: %w", err)
	}

	fmt.Printf("Generated artifacts in %s\n", outDir)
	return nil
}

// injectThreatGroups adds known group definitions to the catalog's Groups
// for any group IDs referenced by threats that aren't already present.
func injectThreatGroups(catalog *gemara.ThreatCatalog) {
	var ids []string
	for _, t := range catalog.Threats {
		ids = append(ids, t.Group)
	}
	injectGroups(&catalog.Groups, ids)
}

// inferThreatCatalogID derives the catalog ID from threat entry IDs by stripping
// the trailing numeric suffix. e.g. "CCC.ObjStor.TH01" → "CCC.ObjStor.TH"
// Core catalogs are mapped to their long canonical IDs.
func inferThreatCatalogID(threats []gemara.Threat) (string, error) {
	if len(threats) == 0 {
		return "", fmt.Errorf("cannot infer catalog ID: threats list is empty")
	}
	short := trailingDigits.ReplaceAllString(threats[0].Id, "")
	return short, nil
}
