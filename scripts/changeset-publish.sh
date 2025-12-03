#!/bin/bash
# Custom publish script for changesets that builds binaries first
set -e

# Get the version from the main package
VERSION=$(node -p "require('./npm/lessgo/package.json').version")
echo "Building and publishing version $VERSION..."

# Build binaries for all platforms
echo "Building binaries..."
./scripts/build-binaries.sh "$VERSION"

# Make Unix binaries executable
chmod +x npm/*/bin/lessc-go 2>/dev/null || true

# Publish all packages using changesets
echo "Publishing packages..."
pnpm changeset publish
