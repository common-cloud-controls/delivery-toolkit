package cmd

import (
	"fmt"
	"strings"

	gemara "github.com/gemaraproj/go-gemara"
	"github.com/gemaraproj/go-gemara/gemaraconv"
)

// sitePageRoot is the path prefix under which the website serves catalog pages.
const sitePageRoot = "/catalogs"

// sitePath builds a site URL like "/catalogs/storage/object/capabilities/v2026.04-rc".
func sitePath(catalogPath, kind, tag string) string {
	return sitePageRoot + "/" + catalogPath + "/" + kind + "/" + tag
}

// siteCrossRefResolver produces gemaraconv.CrossRefResolver that links imports
// and mapped references to sibling pages on the website. It handles the core
// CCC catalogs by name and falls back to a same-service heuristic for refs
// beginning with "CCC." so threats / controls can link to capability pages in
// the same service.
//
// currentPath is the catalogPath being rendered (e.g. "storage/object"); tag is
// the release version (e.g. "v2026.04-rc").
func siteCrossRefResolver(currentPath, tag string) gemaraconv.CrossRefResolver {
	return func(refID, entryID string) string {
		kind, path, ok := siteTargetFor(refID, currentPath)
		if !ok {
			return ""
		}
		return sitePath(path, kind, tag)
	}
}

// siteTargetFor maps a mapping-reference id to (kind, catalogPath).
// Returns ok=false when the reference cannot be resolved to a site page.
func siteTargetFor(refID, currentPath string) (kind, path string, ok bool) {
	switch refID {
	case "CCC.Core.Capabilities", "CCC", "CCC.Core":
		return "capabilities", corePath, true
	case "CCC.Core.Threats":
		return "threats", corePath, true
	case "CCC.Core.Controls":
		return "controls", corePath, true
	}
	// Same-service heuristic: refs like "CCC.ObjStor" refer back to the
	// catalog being rendered. We cannot reliably know the target kind, so we
	// conservatively link to the capabilities page of the same service: that
	// is the typical cross-ref inside a threat or control catalog.
	if strings.HasPrefix(refID, "CCC.") && currentPath != "" {
		return "capabilities", currentPath, true
	}
	return "", "", false
}

// frontmatter builds the YAML frontmatter block that prefixes every generated
// markdown file so the website can render info headers without re-parsing the
// markdown body.
type frontmatter struct {
	Title         string
	Path          string
	CatalogType   string // "capability", "threat", "control"
	Service       string
	Version       string
	GemaraVersion string
	Date          string
	Description   string
	Draft         bool
}

func (f frontmatter) render() string {
	var b strings.Builder
	b.WriteString("---\n")
	writeYAMLField(&b, "title", f.Title)
	writeYAMLField(&b, "path", f.Path)
	writeYAMLField(&b, "catalog_type", f.CatalogType)
	if f.Service != "" {
		writeYAMLField(&b, "service", f.Service)
	}
	if f.Version != "" {
		writeYAMLField(&b, "version", f.Version)
	}
	if f.GemaraVersion != "" {
		writeYAMLField(&b, "gemara_version", f.GemaraVersion)
	}
	if f.Date != "" {
		writeYAMLField(&b, "date", f.Date)
	}
	if f.Description != "" {
		writeYAMLField(&b, "description", f.Description)
	}
	if f.Draft {
		b.WriteString("draft: true\n")
	}
	b.WriteString("---\n\n")
	return b.String()
}

// writeYAMLField writes a `key: "value"` line, double-quoting the value and
// escaping any embedded quotes / backslashes so the file parses as YAML.
func writeYAMLField(b *strings.Builder, key, value string) {
	escaped := strings.ReplaceAll(value, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	fmt.Fprintf(b, "%s: \"%s\"\n", key, escaped)
}

// catalogFrontmatter builds a frontmatter block from a catalog's metadata and
// site coordinates. kind must be one of "capability", "threat", "control".
func catalogFrontmatter(meta gemara.Metadata, title, catalogPath, serviceTitle, tag, kind string) frontmatter {
	return frontmatter{
		Title:         title,
		Path:          sitePath(catalogPath, pluralKind(kind), tag),
		CatalogType:   kind,
		Service:       serviceTitle,
		Version:       tag,
		GemaraVersion: meta.GemaraVersion,
		Date:          string(meta.Date),
		Description:   meta.Description,
		Draft:         meta.Draft,
	}
}

// pluralKind maps a catalog kind to its URL path segment.
func pluralKind(kind string) string {
	switch kind {
	case "capability":
		return "capabilities"
	case "threat":
		return "threats"
	case "control":
		return "controls"
	}
	return kind
}
