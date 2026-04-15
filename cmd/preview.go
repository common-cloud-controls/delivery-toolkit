package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// PreviewLocalCmd generates a catalog and writes its artifacts directly into a
// local checkout of the CCC website repo so `npm run dev` renders the output
// without going through GitHub or a publish step.
var PreviewLocalCmd = &cobra.Command{
	Use:   "preview-local <kind> <path> <title>",
	Short: "Generate and stage artifacts into a local website checkout for preview",
	Long: `Generates <kind> artifacts (capabilities | threats | controls | all) and
copies the resulting markdown + YAML directly into the sibling website repo at
the paths its Vite plugins expect:

  src/content/catalogs/<path>/<kind>/<tag>.md
  public/data/catalogs/<path>/<kind>/<tag>.yaml

Then start the website dev server (` + "`npm run dev`" + `) to preview.

The source catalog is loaded from disk (via --*-dir flags, same as generate) or
fetched from GitHub when no source dir is given.`,
	Args: cobra.ExactArgs(3),
	RunE: runPreviewLocal,
}

func init() {
	PreviewLocalCmd.Flags().String("website-dir", "../common-cloud-controls.github.io", "Path to the website repo checkout")
	PreviewLocalCmd.Flags().String("capabilities-dir", "", "Root of the capability-catalogs repo (omit to fetch from GitHub)")
	PreviewLocalCmd.Flags().String("threats-dir", "", "Root of the threat-catalogs repo (omit to fetch from GitHub)")
	PreviewLocalCmd.Flags().String("controls-dir", "", "Root of the control-catalogs repo (omit to fetch from GitHub)")
	PreviewLocalCmd.Flags().StringP("output-dir", "o", "artifacts", "Directory for intermediate generated files")
	PreviewLocalCmd.Flags().String("tag", "dev", "Release tag used as both the version and the output filename (e.g. v2026.04-rc)")
}

func runPreviewLocal(cmd *cobra.Command, args []string) error {
	kind := args[0]
	catalogPath := args[1]
	titleSuffix := args[2]

	websiteDir, _ := cmd.Flags().GetString("website-dir")
	capsDir, _ := cmd.Flags().GetString("capabilities-dir")
	threatsDir, _ := cmd.Flags().GetString("threats-dir")
	controlsDir, _ := cmd.Flags().GetString("controls-dir")
	outDir, _ := cmd.Flags().GetString("output-dir")
	tag, _ := cmd.Flags().GetString("tag")

	absWebsite, err := filepath.Abs(websiteDir)
	if err != nil {
		return fmt.Errorf("resolving website-dir: %w", err)
	}
	if _, err := os.Stat(filepath.Join(absWebsite, "package.json")); err != nil {
		return fmt.Errorf("website-dir %q does not look like a website checkout (no package.json): %w", absWebsite, err)
	}

	kinds := []string{kind}
	if kind == "all" {
		kinds = []string{"capabilities", "threats", "controls"}
	}

	for _, k := range kinds {
		if err := generateKind(k, catalogPath, titleSuffix, capsDir, threatsDir, controlsDir, outDir, tag); err != nil {
			return err
		}
		if err := stageIntoWebsite(absWebsite, catalogPath, k, outDir, tag); err != nil {
			return err
		}
	}

	fmt.Printf("\nStaged preview at %s\n", absWebsite)
	fmt.Println("Run the website dev server with:")
	fmt.Printf("  cd %s && npm run dev\n", absWebsite)
	return nil
}

// generateKind dispatches to the existing generator for one catalog kind.
func generateKind(kind, catalogPath, titleSuffix, capsDir, threatsDir, controlsDir, outDir, tag string) error {
	switch kind {
	case "capabilities":
		return doGenerateCapabilities(catalogPath, "CCC "+titleSuffix+" Capabilities", titleSuffix, capsDir, outDir, tag)
	case "threats":
		return doGenerateThreats(catalogPath, "CCC "+titleSuffix+" Threats", titleSuffix, threatsDir, outDir, tag)
	case "controls":
		return doGenerateControls(catalogPath, "CCC "+titleSuffix+" Controls", titleSuffix, controlsDir, outDir, tag)
	}
	return fmt.Errorf("unknown kind %q (want capabilities, threats, controls, or all)", kind)
}

// stageIntoWebsite copies the markdown and YAML the generator just wrote into
// the paths the website's Vite plugins read from, using <tag>.md / <tag>.yaml
// as the versioned filename.
func stageIntoWebsite(websiteDir, catalogPath, kind, outDir, tag string) error {
	src := filepath.Join(outDir, catalogPath)

	mdDest := filepath.Join(websiteDir, "src", "content", "catalogs", catalogPath, kind, tag+".md")
	yamlDest := filepath.Join(websiteDir, "public", "data", "catalogs", catalogPath, kind, tag+".yaml")

	if err := copyFile(filepath.Join(src, kind+".md"), mdDest); err != nil {
		return fmt.Errorf("staging markdown: %w", err)
	}
	if err := copyFile(filepath.Join(src, kind+".yaml"), yamlDest); err != nil {
		return fmt.Errorf("staging yaml: %w", err)
	}
	fmt.Printf("Staged %s → %s\n", kind+".md", mdDest)
	fmt.Printf("Staged %s → %s\n", kind+".yaml", yamlDest)
	return nil
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
