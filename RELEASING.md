# Releasing less.go

This project uses [Changesets](https://github.com/changesets/changesets) to manage versioning and releases across all npm packages.

## Overview

All packages are versioned together (fixed versioning):
- `lessgo` - Main package
- `@lessgo/darwin-arm64` - macOS ARM64 binary
- `@lessgo/darwin-x64` - macOS x64 binary
- `@lessgo/linux-x64` - Linux x64 binary
- `@lessgo/linux-arm64` - Linux ARM64 binary
- `@lessgo/win32-x64` - Windows x64 binary
- `@lessgo/win32-arm64` - Windows ARM64 binary

## How to Release

### 1. Create a Changeset

After making changes, create a changeset to describe what changed:

```bash
pnpm changeset
```

This will prompt you to:
1. Select the type of change: `patch` (0.0.x), `minor` (0.x.0), or `major` (x.0.0)
2. Write a summary of the changes

A markdown file will be created in `.changeset/` and automatically committed.

### 2. Open a Pull Request

Push your branch and open a PR as usual. The PR should include:
- Your code changes
- The changeset file (`.changeset/[random-name].md`)

### 3. Merge to Master

When your PR is merged to `master`, the Release workflow runs and:
- Detects the changeset files
- Creates/updates a **"chore: release packages"** PR that:
  - Bumps version numbers in all `package.json` files
  - Updates `CHANGELOG.md` files
  - Deletes the consumed changeset files

### 4. Merge the Release PR

When you're ready to release, merge the "chore: release packages" PR. This triggers:
1. **Build**: Go binaries are cross-compiled for all 6 platforms
2. **Publish**: All packages are published to npm with provenance
3. **Release**: A GitHub Release is created with the changelog

## Commands

| Command | Description |
|---------|-------------|
| `pnpm changeset` | Create a new changeset |
| `pnpm changeset:version` | Apply changesets and bump versions (CI only) |
| `pnpm changeset:publish` | Build binaries and publish to npm (CI only) |

## Changeset Examples

### Patch Release (bug fix)
```
pnpm changeset
# Select: patch
# Summary: "Fix compression flag not being applied"
```

### Minor Release (new feature)
```
pnpm changeset
# Select: minor
# Summary: "Add stdin support for piped input"
```

### Major Release (breaking change)
```
pnpm changeset
# Select: major
# Summary: "Change default math mode to parens-division"
```

## Batching Releases

You can merge multiple PRs with changesets before releasing. They accumulate in the release PR:

```
PR #1: Add feature A (minor)     → merged
PR #2: Fix bug B (patch)         → merged
PR #3: Add feature C (minor)     → merged

Release PR now contains all three changes.
Merging it releases v0.2.0 with all changes.
```

The highest bump type wins: if any changeset is `minor`, the release is `minor`.

## Manual Release (Emergency)

If you need to release manually:

```bash
# Build binaries
./scripts/build-binaries.sh 1.2.3

# Publish (requires npm auth)
pnpm changeset:publish
```

## Configuration

- `.changeset/config.json` - Changesets configuration
- `.github/workflows/release.yml` - Release workflow
- `scripts/changeset-publish.sh` - Custom publish script that builds binaries

## Troubleshooting

### "No changesets found"
You need to create a changeset before merging: `pnpm changeset`

### Release PR not created
Check that:
1. The PR was merged to `master` (not another branch)
2. The PR contained changeset files in `.changeset/`
3. The Release workflow ran successfully

### npm publish failed
Check that:
1. npm trusted publishing is configured for all packages
2. The workflow has `id-token: write` permission
