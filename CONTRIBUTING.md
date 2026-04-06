# Contributing to the CCC Delivery Toolkit

The delivery toolkit is the `ccc` CLI — a Go binary that generates, releases, and publishes CCC catalog artifacts. This guide covers how to work on the toolkit and how the release pipeline works end to end.

---

## Development

### Prerequisites

- Go 1.21+
- Access to the `common-cloud-controls` GitHub org

### Build

```sh
go build -o ccc .
```

### Run locally

```sh
# Generate artifacts from a local catalog repo
./ccc generate capabilities storage/object "Object Storage" \
  --capabilities-dir ../capability-catalogs \
  --output-dir ./out

# Generate all types for a service
./ccc generate capabilities storage/object "Object Storage" --capabilities-dir ../capability-catalogs
./ccc generate threats storage/object "Object Storage" --threats-dir ../threat-catalogs
./ccc generate controls storage/object "Object Storage" --controls-dir ../control-catalogs

# Fetch directly from GitHub (omit the --*-dir flag)
./ccc generate capabilities storage/object "Object Storage"
```

The core catalog uses the special path `core/ccc`:

```sh
./ccc generate all core/ccc "Core"
```

---

## Command Reference

| Command | Description |
|---|---|
| `ccc generate <type> <path> <title>` | Generate artifacts without publishing |
| `ccc release <type> <path> <title>` | Generate artifacts for release (identical to generate for now; exists as a distinct CI step) |
| `ccc release all <path> <title>` | Release capabilities, threats, and controls in sequence |
| `ccc publish <type> <path>` | Publish a single artifact type to the website repository |
| `ccc publish all <path>` | Publish all artifact types to the website repository |

`<type>` is one of: `capabilities`, `threats`, `controls`, `all`

`<path>` is the catalog path within the relevant repo (e.g. `storage/object`, `key-management/kms`, `core/ccc`)

---

## How Releases Work

Catalog releases flow through two sequential steps: **release** and **publish**. Keeping them separate means CI fails fast with a clear signal before any changes reach the website.

### Step 1 — `ccc release`

Reads a catalog YAML from the relevant source repo (or from GitHub if no local path is given), injects standard CCC metadata, and writes two artifacts per type to `--output-dir`:

- `<path>/capabilities.yaml` — the populated YAML artifact
- `<path>/capabilities.md` — a rendered Markdown table

This step must succeed before anything is published. If the source YAML is malformed, incomplete, or fails to render, the pipeline stops here.

### Step 2 — `ccc publish`

Reads the artifacts written by `ccc release` and commits them to the [website repository](https://github.com/common-cloud-controls/website-new) via the GitHub Contents API.

Each artifact type produces two files in the website repo:

| Artifact | Website path | Purpose |
|---|---|---|
| `capabilities.md` (with frontmatter) | `src/content/catalogs/<path>/<tag>-capabilities.md` | Rendered page at `/catalogs/<path>/<tag>-capabilities` |
| `capabilities.yaml` | `public/data/catalogs/<path>/<tag>-capabilities.yaml` | Raw data file served as a static URL |

The same pattern applies for `threats` and `controls`.

The commit to the website repo triggers its GitHub Actions Pages workflow automatically, deploying the new pages within minutes.

### Authentication

`ccc publish` requires a GitHub token with write access to the website repository. Pass it via flag or environment variable:

```sh
# Via flag
./ccc publish all storage/object --tag v1.0.0 --website-repo common-cloud-controls/website-new --token ghp_...

# Via environment variable (preferred in CI)
export GITHUB_TOKEN=ghp_...
./ccc publish all storage/object --tag v1.0.0 --website-repo common-cloud-controls/website-new
```

---

## CI Workflow

Each catalog repository (capability-catalogs, threat-catalogs, control-catalogs, core-catalog) should have a release workflow that runs on tag push. A minimal example:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install ccc
        run: |
          go install github.com/finos/common-cloud-controls@latest
          # or: download a pre-built binary from the delivery-toolkit releases

      - name: Release artifacts
        run: |
          ccc release all storage/object "Object Storage"

      - name: Publish to website
        run: |
          ccc publish all storage/object \
            --tag ${{ github.ref_name }} \
            --website-repo common-cloud-controls/website-new
        env:
          GITHUB_TOKEN: ${{ secrets.WEBSITE_PAT }}
```

`WEBSITE_PAT` is a fine-grained personal access token (or GitHub App token) stored as a secret in the catalog repo, scoped to `Contents: read & write` on the website repository.

---

## Where Artifacts End Up

After a successful publish, artifacts are available at:

- **Rendered page:** `https://common-cloud-controls.github.io/catalogs/<path>/<tag>-<type>`
  e.g. `https://common-cloud-controls.github.io/catalogs/storage/object/v1.0.0-controls`

- **Raw YAML:** `https://common-cloud-controls.github.io/data/catalogs/<path>/<tag>-<type>.yaml`
  e.g. `https://common-cloud-controls.github.io/data/catalogs/storage/object/v1.0.0-controls.yaml`

The raw YAML URLs are stable and can be referenced by downstream tooling, compliance pipelines, or other automation.
