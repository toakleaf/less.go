#!/bin/bash
# Atomic release script: version + publish in one step
# Only commits version bumps after successful npm publish
# Uses npm publish directly for OIDC trusted publisher support
set -e

echo "=== Starting atomic release process ==="
echo "Node version: $(node --version)"
echo "npm version: $(npm --version)"

# Check if there are changesets to release
CHANGESETS=$(ls -A .changeset/*.md 2>/dev/null | grep -v README.md || true)
if [ -z "$CHANGESETS" ]; then
  echo "No changesets found. Nothing to release."
  echo "released=false" >> $GITHUB_OUTPUT
  exit 0
fi

echo "Found changesets:"
echo "$CHANGESETS"
echo ""

# Save current state so we can revert on failure
git stash push -m "pre-release-stash" --include-untracked || true

# Run changeset version (updates package.json files and changelogs)
echo "Running changeset version..."
pnpm changeset version

# Update lockfile
echo "Updating lockfile..."
pnpm install --no-frozen-lockfile

VERSION=$(node -p "require('./npm/lessgo/package.json').version")
PLUGIN_VITE_VERSION=$(node -p "require('./packages/plugin-vite/package.json').version")
echo "New lessgo version: $VERSION"
echo "New plugin-vite version: $PLUGIN_VITE_VERSION"

# Build binaries for all platforms
echo ""
echo "Building binaries..."
./scripts/build-binaries.sh "$VERSION"

# Make Unix binaries executable
chmod +x npm/*/bin/lessc-go 2>/dev/null || true

# Build Vite plugin
echo ""
echo "Building Vite plugin..."
pnpm --filter @lessgo/plugin-vite run build

# Platform packages must be published first (lessgo depends on them as optionalDependencies)
PLATFORM_PACKAGES=(
  "darwin-arm64"
  "darwin-x64"
  "linux-x64"
  "linux-arm64"
  "win32-x64"
  "win32-arm64"
)

publish_failed() {
  echo ""
  echo "=== PUBLISH FAILED ==="
  echo "Reverting version changes..."
  git checkout -- .
  git clean -fd
  git stash pop 2>/dev/null || true
  exit 1
}

# =============================================================================
# DRY-RUN PHASE: Verify all packages can be published before publishing any
# This prevents partial releases where some packages are published but others fail
# =============================================================================
echo ""
echo "=== DRY-RUN: Verifying all packages can be published ==="

DRY_RUN_FAILED=false

for pkg in "${PLATFORM_PACKAGES[@]}"; do
  echo "Verifying @lessgo/$pkg..."
  if ! npm publish "./npm/$pkg" --access public --dry-run 2>&1; then
    echo "ERROR: Dry-run failed for @lessgo/$pkg"
    DRY_RUN_FAILED=true
  fi
done

echo "Verifying lessgo..."
if ! npm publish "./npm/lessgo" --access public --dry-run 2>&1; then
  echo "ERROR: Dry-run failed for lessgo"
  DRY_RUN_FAILED=true
fi

# Pack plugin-vite with pnpm to resolve workspace:* protocol
echo "Packing @lessgo/plugin-vite (resolves workspace:* to actual version)..."
# pnpm pack returns the full absolute path, so we extract just the filename
PLUGIN_VITE_TARBALL_PATH=$(cd packages/plugin-vite && pnpm pack --pack-destination ../../ 2>/dev/null | tail -1)
PLUGIN_VITE_TARBALL=$(basename "$PLUGIN_VITE_TARBALL_PATH")
if [ -z "$PLUGIN_VITE_TARBALL" ]; then
  echo "ERROR: Failed to pack @lessgo/plugin-vite"
  DRY_RUN_FAILED=true
else
  echo "Verifying @lessgo/plugin-vite..."
  if ! npm publish "./$PLUGIN_VITE_TARBALL" --access public --dry-run 2>&1; then
    echo "ERROR: Dry-run failed for @lessgo/plugin-vite"
    DRY_RUN_FAILED=true
  fi
fi

if [ "$DRY_RUN_FAILED" = true ]; then
  echo ""
  echo "=== DRY-RUN FAILED ==="
  echo "One or more packages failed validation. No packages were published."
  echo "Fix the issues above and try again."
  publish_failed
fi

echo ""
echo "=== DRY-RUN PASSED: All packages validated successfully ==="
echo ""

# =============================================================================
# PUBLISH PHASE: Actually publish packages (only runs if dry-run passed)
# =============================================================================
echo "Publishing packages with npm (OIDC trusted publishers)..."

for pkg in "${PLATFORM_PACKAGES[@]}"; do
  echo ""
  echo "=== Publishing @lessgo/$pkg ==="
  if ! npm publish "./npm/$pkg" --access public; then
    echo "ERROR: Failed to publish @lessgo/$pkg"
    echo "WARNING: Some packages may have been published. Manual intervention may be required."
    publish_failed
  fi
done

echo ""
echo "=== Publishing lessgo ==="
if ! npm publish "./npm/lessgo" --access public; then
  echo "ERROR: Failed to publish lessgo"
  echo "WARNING: Some packages may have been published. Manual intervention may be required."
  publish_failed
fi

# Publish Vite plugin using the tarball created during dry-run
# (pnpm pack resolves workspace:* to actual version)
echo ""
echo "=== Publishing @lessgo/plugin-vite ==="
if [ -z "$PLUGIN_VITE_TARBALL" ] || [ ! -f "./$PLUGIN_VITE_TARBALL" ]; then
  echo "ERROR: Plugin-vite tarball not found. Re-packing..."
  PLUGIN_VITE_TARBALL_PATH=$(cd packages/plugin-vite && pnpm pack --pack-destination ../../ 2>/dev/null | tail -1)
  PLUGIN_VITE_TARBALL=$(basename "$PLUGIN_VITE_TARBALL_PATH")
fi
if ! npm publish "./$PLUGIN_VITE_TARBALL" --access public; then
  echo "ERROR: Failed to publish @lessgo/plugin-vite"
  echo "WARNING: Some packages may have been published. Manual intervention may be required."
  publish_failed
fi
# Clean up tarball
rm -f "./$PLUGIN_VITE_TARBALL"

# Success! Now commit the version changes
echo ""
echo "=== All packages published successfully! ==="
echo "Committing version changes..."

# Build list of published packages
PUBLISHED_PACKAGES="- lessgo@$VERSION
- @lessgo/darwin-arm64@$VERSION
- @lessgo/darwin-x64@$VERSION
- @lessgo/linux-x64@$VERSION
- @lessgo/linux-arm64@$VERSION
- @lessgo/win32-x64@$VERSION
- @lessgo/win32-arm64@$VERSION
- @lessgo/plugin-vite@$PLUGIN_VITE_VERSION"

git add -A
git commit -m "chore: release v$VERSION

Published packages:
$PUBLISHED_PACKAGES"

# Clean up stash
git stash drop 2>/dev/null || true

# Signal successful release to GitHub Actions
echo "released=true" >> $GITHUB_OUTPUT
echo "version=$VERSION" >> $GITHUB_OUTPUT

echo ""
echo "=== Release v$VERSION complete! ==="
