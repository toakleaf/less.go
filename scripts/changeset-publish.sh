#!/bin/bash
# Custom publish script for changesets that builds binaries first
# Uses npm publish directly for OIDC trusted publisher support (pnpm doesn't support OIDC yet)
set -e

# Get the version from the main package
VERSION=$(node -p "require('./npm/lessgo/package.json').version")
echo "Building and publishing version $VERSION..."

# Build binaries for all platforms
echo "Building binaries..."
./scripts/build-binaries.sh "$VERSION"

# Make Unix binaries executable
chmod +x npm/*/bin/lessc-go 2>/dev/null || true

# Check if this version is already published
PUBLISHED_VERSION=$(npm view lessgo version 2>/dev/null || echo "0.0.0")
if [ "$PUBLISHED_VERSION" = "$VERSION" ]; then
  echo "Version $VERSION is already published, skipping..."
  exit 0
fi

echo "Publishing packages with npm (for OIDC trusted publisher support)..."

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

echo "Successfully published version $VERSION"
