# Contributing to the CCC Delivery Toolkit

The delivery toolkit is the `ccc` CLI — a Go binary that generates, releases, and publishes CCC catalog artifacts.

---

## Development

**Prerequisites:** Go 1.21+, access to the `common-cloud-controls` GitHub org.

```sh
go build -o ccc .
```

### Running locally

```sh
# Generate from a local catalog repo
./ccc generate capabilities storage/object "Object Storage" --capabilities-dir ../capability-catalogs

# Fetch from GitHub (omit --*-dir)
./ccc generate capabilities storage/object "Object Storage"
```

The core catalog uses the reserved path `core/ccc`:

```sh
./ccc generate all core/ccc "Core"
```

---

## Commands

```
ccc (generate|release|publish) (capabilities|threats|controls|all) <path> [<title>]
```

| Command | Description |
|---|---|
| `generate` | Build artifacts locally — no publishing |
| `release` | Same as generate; exists as a distinct CI step so the pipeline fails before publish if artifacts can't be built |
| `publish` | Commit built artifacts to the website repository |

**Arguments**

- `<path>` — catalog path within the source repo, e.g. `storage/object`, `core/ccc`
- `<title>` — human-readable service name, e.g. `"Object Storage"` (required by `generate` and `release`, not by `publish`)

**Common flags**

| Flag | Commands | Description |
|---|---|---|
| `--output-dir` | all | Artifact directory (default: `artifacts`) |
| `--capabilities-dir` | `generate`, `release` | Local capability-catalogs root |
| `--threats-dir` | `generate`, `release` | Local threat-catalogs root |
| `--controls-dir` | `generate`, `release` | Local control-catalogs root |
| `--tag` | `publish` | Release tag, e.g. `v1.0.0` (required) |
| `--website-repo` | `publish` | Target repo in `owner/repo` format (required) |
| `--token` | `publish` | GitHub token; falls back to `$GITHUB_TOKEN` |

---

## Release Pipeline

Releases run in two sequential steps. If `release` fails, `publish` never runs.

### 1. `ccc release`

Reads a catalog YAML, injects CCC metadata, and writes to `--output-dir`:

```
<path>/capabilities.yaml   # populated YAML artifact
<path>/capabilities.md     # rendered Markdown table
```

### 2. `ccc publish`

Commits the built artifacts to the website repo via the GitHub Contents API. Each type produces two files:

```
src/content/catalogs/<path>/<tag>-<type>.md     # becomes a rendered page
public/data/catalogs/<path>/<tag>-<type>.yaml   # served as a static file
```

The commit triggers the website's Pages workflow automatically.

### CI example

```yaml
on:
  push:
    tags: ['v*']

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Release artifacts
        run: ccc release all storage/object "Object Storage"
      - name: Publish to website
        run: |
          ccc publish all storage/object \
            --tag ${{ github.ref_name }} \
            --website-repo common-cloud-controls/website-new
        env:
          GITHUB_TOKEN: ${{ secrets.WEBSITE_PAT }}
```

`WEBSITE_PAT` is a fine-grained PAT scoped to `Contents: read & write` on the website repo.

---

## Published URLs

```
# Rendered page
https://common-cloud-controls.github.io/catalogs/<path>/<tag>-<type>

# Raw YAML (stable, safe to reference from tooling)
https://common-cloud-controls.github.io/data/catalogs/<path>/<tag>-<type>.yaml
```
