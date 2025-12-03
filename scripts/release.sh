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
echo "New version: $VERSION"

# Build binaries for all platforms
echo ""
echo "Building binaries..."
./scripts/build-binaries.sh "$VERSION"

# Make Unix binaries executable
chmod +x npm/*/bin/lessc-go 2>/dev/null || true

echo ""
echo "Publishing packages with npm (OIDC trusted publishers)..."

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

for pkg in "${PLATFORM_PACKAGES[@]}"; do
  echo ""
  echo "=== Publishing @lessgo/$pkg ==="
  if ! npm publish "./npm/$pkg" --access public; then
    echo "ERROR: Failed to publish @lessgo/$pkg"
    publish_failed
  fi
done

echo ""
echo "=== Publishing lessgo ==="
if ! npm publish "./npm/lessgo" --access public; then
  echo "ERROR: Failed to publish lessgo"
  publish_failed
fi

# Success! Now commit the version changes
echo ""
echo "=== All packages published successfully! ==="
echo "Committing version changes..."

git add -A
git commit -m "chore: release v$VERSION

Published packages:
- lessgo@$VERSION
- @lessgo/darwin-arm64@$VERSION
- @lessgo/darwin-x64@$VERSION
- @lessgo/linux-x64@$VERSION
- @lessgo/linux-arm64@$VERSION
- @lessgo/win32-x64@$VERSION
- @lessgo/win32-arm64@$VERSION"

# Clean up stash
git stash drop 2>/dev/null || true

echo ""
echo "=== Release v$VERSION complete! ==="
