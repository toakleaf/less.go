#!/bin/bash
# Atomic release script: only commits version bumps after successful npm publish
# Uses npm publish directly for OIDC trusted publisher support
set -e

echo "=== Starting atomic release process ==="

# Check if there are changesets to release
if [ ! -d ".changeset" ] || [ -z "$(ls -A .changeset/*.md 2>/dev/null | grep -v README.md)" ]; then
  echo "No changesets found, checking if versions need publishing..."

  # Check if current version is already published
  VERSION=$(node -p "require('./npm/lessgo/package.json').version")
  PUBLISHED_VERSION=$(npm view lessgo version 2>/dev/null || echo "0.0.0")

  if [ "$PUBLISHED_VERSION" = "$VERSION" ]; then
    echo "Version $VERSION is already published. Nothing to do."
    exit 0
  fi

  echo "Version $VERSION not yet published, proceeding with publish..."
else
  echo "Found changesets, running version bump..."

  # Run changeset version (updates package.json files but we won't commit yet)
  pnpm changeset version

  # Update lockfile
  pnpm install --no-frozen-lockfile

  VERSION=$(node -p "require('./npm/lessgo/package.json').version")
  echo "Bumped to version $VERSION"
fi

# Build binaries for all platforms
echo "Building binaries..."
./scripts/build-binaries.sh "$VERSION"

# Make Unix binaries executable
chmod +x npm/*/bin/lessc-go 2>/dev/null || true

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

PUBLISH_FAILED=0

for pkg in "${PLATFORM_PACKAGES[@]}"; do
  echo "Publishing @lessgo/$pkg..."
  if ! npm publish "./npm/$pkg" --access public; then
    echo "ERROR: Failed to publish @lessgo/$pkg"
    PUBLISH_FAILED=1
    break
  fi
done

if [ $PUBLISH_FAILED -eq 0 ]; then
  echo "Publishing lessgo..."
  if ! npm publish "./npm/lessgo" --access public; then
    echo "ERROR: Failed to publish lessgo"
    PUBLISH_FAILED=1
  fi
fi

if [ $PUBLISH_FAILED -eq 1 ]; then
  echo ""
  echo "=== PUBLISH FAILED ==="
  echo "Some packages may have been published. Manual intervention may be required."
  echo "Reverting local version changes..."
  git checkout -- .
  exit 1
fi

echo ""
echo "=== All packages published successfully! ==="
echo "Committing version changes..."

# Now commit the version bumps since publish succeeded
git add -A
git commit -m "chore: release v$VERSION" || echo "Nothing to commit"

echo "Version $VERSION released successfully!"
echo "published=true" >> "${GITHUB_OUTPUT:-/dev/null}"
