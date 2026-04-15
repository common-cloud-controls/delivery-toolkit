package cmd

import (
	"strings"
	"testing"

	gemara "github.com/gemaraproj/go-gemara"
)

func TestSiteCrossRefResolver_coreCatalogs(t *testing.T) {
	r := siteCrossRefResolver("storage/object", "v2026.04-rc")
	cases := []struct {
		refID, entryID, want string
	}{
		{"CCC.Core.Capabilities", "CCC.Core.CP01", "/catalogs/core/ccc/capabilities/v2026.04-rc"},
		{"CCC.Core.Threats", "", "/catalogs/core/ccc/threats/v2026.04-rc"},
		{"CCC.Core.Controls", "CCC.Core.CN01", "/catalogs/core/ccc/controls/v2026.04-rc"},
		{"CCC", "", "/catalogs/core/ccc/capabilities/v2026.04-rc"},
	}
	for _, c := range cases {
		got := r(c.refID, c.entryID)
		if got != c.want {
			t.Errorf("resolver(%q, %q) = %q, want %q", c.refID, c.entryID, got, c.want)
		}
	}
}

func TestSiteCrossRefResolver_sameService(t *testing.T) {
	r := siteCrossRefResolver("storage/object", "v2026.04-rc")
	got := r("CCC.ObjStor", "CCC.ObjStor.CP01")
	want := "/catalogs/storage/object/capabilities/v2026.04-rc"
	if got != want {
		t.Errorf("same-service resolver = %q, want %q", got, want)
	}
}

func TestSiteCrossRefResolver_unknown(t *testing.T) {
	r := siteCrossRefResolver("storage/object", "v2026.04-rc")
	if got := r("Unknown", "entry"); got != "" {
		t.Errorf("unknown ref should not resolve, got %q", got)
	}
}

func TestFrontmatter_render(t *testing.T) {
	fm := frontmatter{
		Title:         "CCC Object Storage Capabilities",
		Path:          "/catalogs/storage/object/capabilities/v2026.04-rc",
		CatalogType:   "capability",
		Service:       "Object Storage",
		Version:       "v2026.04-rc",
		GemaraVersion: "v1.0.0",
		Date:          "2026-04-15",
		Description:   "Capabilities for \"Object Storage\".",
		Draft:         true,
	}
	got := fm.render()

	wantLines := []string{
		`---`,
		`title: "CCC Object Storage Capabilities"`,
		`path: "/catalogs/storage/object/capabilities/v2026.04-rc"`,
		`catalog_type: "capability"`,
		`service: "Object Storage"`,
		`version: "v2026.04-rc"`,
		`gemara_version: "v1.0.0"`,
		`date: "2026-04-15"`,
		`description: "Capabilities for \"Object Storage\"."`,
		`draft: true`,
		`---`,
	}
	for _, want := range wantLines {
		if !strings.Contains(got, want) {
			t.Errorf("frontmatter missing line %q\n---\n%s", want, got)
		}
	}
	// Blank line between frontmatter and body
	if !strings.HasSuffix(got, "---\n\n") {
		t.Errorf("frontmatter must end with blank line; got suffix %q", got[len(got)-10:])
	}
}

func TestCatalogFrontmatter_fromMetadata(t *testing.T) {
	meta := gemara.Metadata{
		GemaraVersion: "v1.0.0",
		Date:          gemara.Datetime("2026-04-15"),
		Description:   "Capabilities for Object Storage.",
		Draft:         false,
	}
	fm := catalogFrontmatter(meta, "CCC Object Storage Capabilities", "storage/object", "Object Storage", "v2026.04-rc", "capability")

	if fm.Path != "/catalogs/storage/object/capabilities/v2026.04-rc" {
		t.Errorf("wrong path: %s", fm.Path)
	}
	if fm.CatalogType != "capability" {
		t.Errorf("wrong catalog_type: %s", fm.CatalogType)
	}
	if fm.GemaraVersion != "v1.0.0" {
		t.Errorf("gemara version not carried: %s", fm.GemaraVersion)
	}
	if fm.Date != "2026-04-15" {
		t.Errorf("date not carried: %s", fm.Date)
	}
}

func TestPluralKind(t *testing.T) {
	cases := map[string]string{
		"capability": "capabilities",
		"threat":     "threats",
		"control":    "controls",
		"unknown":    "unknown",
	}
	for in, want := range cases {
		if got := pluralKind(in); got != want {
			t.Errorf("pluralKind(%q) = %q, want %q", in, got, want)
		}
	}
}
