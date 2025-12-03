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

A markdown file will be created in `.changeset/`.

### 2. Open a Pull Request

Push your branch and open a PR as usual. The PR should include:
- Your code changes
- The changeset file (`.changeset/[random-name].md`)

### 3. Merge to Master

When your PR is merged to `master`, the Release workflow automatically:

1. **Detects changesets** - Looks for `.changeset/*.md` files
2. **Bumps versions** - Updates all `package.json` and `CHANGELOG.md` files
3. **Builds binaries** - Cross-compiles Go binaries for all 6 platforms
4. **Publishes to npm** - Publishes all packages using OIDC trusted publishing
5. **Commits version bump** - Only if publish succeeds
6. **Creates GitHub Release** - With changelog notes

**Important:** Version bumps are only committed after successful npm publish. If publishing fails, changes are reverted and no version bump is committed.

## Commands

| Command | Description |
|---------|-------------|
| `pnpm changeset` | Create a new changeset |
| `pnpm release` | Run the full release process locally (requires npm auth) |

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

You can merge multiple PRs with changesets before releasing. They accumulate:

```
PR #1: Add feature A (minor)     → merged, triggers release
PR #2: Fix bug B (patch)         → merged, triggers release
PR #3: Add feature C (minor)     → merged, triggers release
```

Each merge with a changeset triggers its own release.

The highest bump type wins: if any changeset is `minor`, the release is `minor`.

## Manual Release (Emergency)

If you need to release manually:

```bash
# Ensure you're logged in to npm
npm login

# Run the release script
pnpm release
```

## Configuration

- `.changeset/config.json` - Changesets configuration
- `.github/workflows/release.yml` - Release workflow
- `scripts/release.sh` - Atomic release script (version + build + publish)

## Troubleshooting

### "No changesets found"
You need to create a changeset before merging: `pnpm changeset`

### Release not triggered
Check that:
1. The PR was merged to `master` (not another branch)
2. The PR contained changeset files in `.changeset/`
3. The Release workflow ran successfully

### npm publish failed
Check that:
1. npm trusted publishing is configured for all packages on npmjs.com
2. The workflow has `id-token: write` permission
3. npm CLI is version 11.5.1+ (for OIDC support)

### Version not bumped after failed publish
This is expected! The release process is atomic - versions are only committed after successful npm publish. If publish fails, run the workflow again after fixing the issue.
