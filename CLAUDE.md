# CLAUDE.md

## Overview

The delivery toolkit for [FINOS Common Cloud Controls](https://commoncloudcontrols.com). A Go CLI (`ccc`) that generates release artifacts (YAML + Markdown) from source catalogs and publishes them to the website repo.

## Commands

```bash
go build -o ccc .          # Build the CLI
go test ./...              # Run tests
go vet ./...               # Static analysis
```

## CLI Structure

```
ccc generate (capabilities|threats|controls|all) <path> <title>
ccc release  (capabilities|threats|controls|all) <path> <title>
ccc publish  (capabilities|threats|controls|all) <path>
```

- **generate** — read source YAML, inject metadata + groups, write enriched YAML + Markdown
- **release** — same as generate (used in CI)
- **publish** — commit generated artifacts to the website repo via GitHub API

## Key Files

- `cmd/groups.go` — canonical group definitions (`knownGroups`). All 12 groups are defined here. When a catalog entry references a group ID, its full definition (title + description) is injected into the output.
- `cmd/generate_capabilities.go` — capabilities generation + group injection
- `cmd/generate_controls.go` — controls generation + group injection
- `cmd/generate_threats.go` — threats generation + group injection
- `cmd/generate_capabilities.go` — shared helpers (`mappingRefsFromImports`, `inferCatalogID`, `resolveLocalPath`, `fetchURL`)
- `cmd/publish.go` — GitHub API helpers for committing to the website repo
- `cmd/markdown.go` — Markdown rendering templates
- `action.yml` — composite GitHub Action used by catalog repo release workflows

## Groups

Group definitions in `cmd/groups.go` are the single source of truth. Available IDs:

`Encryption`, `Access`, `Observability`, `Data`, `Resource`, `Compute`, `Ingestion`, `Networking`, `Orchestration`, `Processing`, `Messaging`, `MachineLearning`

Adding a new group requires updating `knownGroups` in `groups.go`. See the parent repo's CLAUDE.md for the full group assignment guide.

## Source Catalog Format

The toolkit expects catalogs with a flat top-level list:
- `capabilities:` — list of capability entries
- `controls:` — list of control entries (not `control-families:`)
- `threats:` — list of threat entries

Each entry must have `id`, `title`, `group`, and type-specific fields.

## Release Workflow

1. Catalog repo pushes a `v*` tag
2. GitHub Actions workflow builds `ccc` binary
3. For each service in the matrix: `ccc release <type> <path> <title> --<type>-dir . --tag <tag>`
4. Then: `ccc publish <type> <path> --tag <tag> --token <bot-token>`
5. Publish commits enriched YAML + Markdown to `common-cloud-controls.github.io`
