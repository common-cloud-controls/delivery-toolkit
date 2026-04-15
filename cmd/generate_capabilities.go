package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	gemara "github.com/gemaraproj/go-gemara"
	"github.com/gemaraproj/go-gemara/gemaraconv"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var trailingDigits = regexp.MustCompile(`\d+$`)

// coreCatalogTitles maps core catalog IDs to human-readable titles.
var coreCatalogTitles = map[string]string{
	"CCC.Core.Capabilities": "CCC Core Capabilities",
	"CCC.Core.Threats":      "CCC Core Threats",
	"CCC.Core.Controls":     "CCC Core Controls",
}


// mappingRefsFromImports derives MappingReferences from a catalog's Imports
// field. Each unique ReferenceId in the imports produces one MappingReference
// entry so that gemara can resolve cross-catalog references.
func mappingRefsFromImports(imports []gemara.MultiEntryMapping, version string) []gemara.MappingReference {
	seen := map[string]bool{}
	var refs []gemara.MappingReference
	for _, imp := range imports {
		id := imp.ReferenceId
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		title := coreCatalogTitles[id]
		if title == "" {
			title = id
		}
		refs = append(refs, gemara.MappingReference{
			Id:      id,
			Title:   title,
			Version: version,
		})
	}
	return refs
}

// inferCatalogID derives the catalog ID from capability entry IDs by stripping
// the trailing numeric suffix. e.g. "CCC.ObjStor.CP01" → "CCC.ObjStor.CP"
// Core catalogs are mapped to their long canonical IDs.
func inferCatalogID(capabilities []gemara.Capability) (string, error) {
	if len(capabilities) == 0 {
		return "", fmt.Errorf("cannot infer catalog ID: capabilities list is empty")
	}
	short := trailingDigits.ReplaceAllString(capabilities[0].Id, "")
	return short, nil
}

const githubRawBase = "https://raw.githubusercontent.com/common-cloud-controls/capability-catalogs/refs/heads/main"
const githubRawCoreBase = "https://raw.githubusercontent.com/common-cloud-controls/core-catalog/refs/heads/main"

const corePath = "core/ccc"

// resolveGitHubURL returns the full GitHub raw URL for a catalog file.
// The core catalog (path "core/ccc") lives in a separate repo with files at
// the root; all others use the provided repoBase with the path as a subdirectory.
func resolveGitHubURL(repoBase, catalogPath, filename string) string {
	if catalogPath == corePath {
		return githubRawCoreBase + "/" + filename
	}
	return repoBase + "/" + catalogPath + "/" + filename
}

// resolveLocalPath returns the filesystem path to a catalog file.
// For "core/ccc", files are at the root of the core-catalog repo,
// so the caller passes that repo root as dir.
func resolveLocalPath(dir, catalogPath, filename string) string {
	if catalogPath == corePath {
		return filepath.Join(dir, filename)
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
For core/ccc:   ` + githubRawCoreBase + `/capabilities.yaml`,
	Args: cobra.ExactArgs(2),
	RunE: runGenerateCapabilities,
}

func init() {
	generateCapabilitiesCmd.Flags().String("capabilities-dir", "", "Root of the capability-catalogs repo (omit to fetch from GitHub)")
	generateCapabilitiesCmd.Flags().StringP("output-dir", "o", "artifacts", "Directory to write generated files into")
	generateCapabilitiesCmd.Flags().String("tag", "dev", "Release tag to embed in artifact metadata (e.g. v2026.04-rc)")
}

func runGenerateCapabilities(cmd *cobra.Command, args []string) error {
	capabilitiesDir, _ := cmd.Flags().GetString("capabilities-dir")
	outputDir, _ := cmd.Flags().GetString("output-dir")
	tag, _ := cmd.Flags().GetString("tag")
	return doGenerateCapabilities(args[0], "CCC "+args[1]+" Capabilities", args[1], capabilitiesDir, outputDir, tag)
}

func doGenerateCapabilities(catalogPath, catalogTitle, serviceTitle, capabilitiesDir, outputDir, tag string) error {
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
	catalogID, err := inferCatalogID(catalog.Capabilities)
	if err != nil {
		return err
	}

	catalog.Title = catalogTitle
	catalog.Metadata = gemara.Metadata{
		Id:                catalogID,
		Type:              gemara.CapabilityCatalogArtifact,
		GemaraVersion:     gemara.SchemaVersion,
		Version:           tag,
		Description:       "Capabilities for " + serviceTitle + " technologies, as defined by the FINOS Common Cloud Controls project.",
		MappingReferences: mappingRefsFromImports(catalog.Imports, tag),
		Author: gemara.Actor{
			Id:   "FINOS-CCC",
			Name: "FINOS Common Cloud Controls",
			Type: gemara.Human,
		},
	}

	// Inject group definitions for any referenced capability families
	injectCapabilityGroups(&catalog)

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

	// Render markdown via go-gemara and prefix site-friendly frontmatter.
	body, err := gemaraconv.CapabilityCatalog(&catalog).ToMarkdown(
		context.Background(),
		gemaraconv.WithCrossRefResolver(siteCrossRefResolver(catalogPath, tag)),
	)
	if err != nil {
		return fmt.Errorf("rendering Markdown: %w", err)
	}
	fm := catalogFrontmatter(catalog.Metadata, catalogTitle, catalogPath, serviceTitle, tag, "capability")
	out := append([]byte(fm.render()), body...)
	if err := os.WriteFile(filepath.Join(outDir, "capabilities.md"), out, 0644); err != nil {
		return fmt.Errorf("writing capabilities.md: %w", err)
	}

	fmt.Printf("Generated artifacts in %s\n", outDir)
	return nil
}

// injectCapabilityGroups adds known group definitions to the catalog's Groups
// for any group IDs referenced by capabilities that aren't already present.
func injectCapabilityGroups(catalog *gemara.CapabilityCatalog) {
	var ids []string
	for _, c := range catalog.Capabilities {
		ids = append(ids, c.Group)
	}
	injectGroups(&catalog.Groups, ids)
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
