#!/bin/bash
# Publish script: builds and publishes packages using npm for OIDC support
# Called by changesets action AFTER version PR is merged
set -e

echo "=== Starting publish process ==="

# Get the version from the main package
VERSION=$(node -p "require('./npm/lessgo/package.json').version")

# Check if this version is already published
PUBLISHED_VERSION=$(npm view lessgo version 2>/dev/null || echo "0.0.0")
if [ "$PUBLISHED_VERSION" = "$VERSION" ]; then
  echo "Version $VERSION is already published. Nothing to do."
  exit 0
fi

echo "Publishing version $VERSION..."

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

for pkg in "${PLATFORM_PACKAGES[@]}"; do
  echo "Publishing @lessgo/$pkg..."
  npm publish "./npm/$pkg" --access public
done

# Publish main package last
echo "Publishing lessgo..."
npm publish "./npm/lessgo" --access public

echo ""
echo "=== Successfully published version $VERSION ==="
