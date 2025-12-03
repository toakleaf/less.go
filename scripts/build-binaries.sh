#!/bin/bash
# Cross-compile lessc-go for all supported platforms

set -e

VERSION=${1:-"dev"}
OUTPUT_DIR="npm"
RUNTIME_DIR="less/runtime"

# Platform targets: GOOS/GOARCH -> npm package directory
declare -A TARGETS=(
  ["darwin/arm64"]="darwin-arm64"
  ["darwin/amd64"]="darwin-x64"
  ["linux/amd64"]="linux-x64"
  ["linux/arm64"]="linux-arm64"
  ["windows/amd64"]="win32-x64"
  ["windows/arm64"]="win32-arm64"
)

for target in "${!TARGETS[@]}"; do
  GOOS="${target%/*}"
  GOARCH="${target#*/}"
  SUFFIX="${TARGETS[$target]}"

  OUTPUT_NAME="lessc-go"
  if [ "$GOOS" = "windows" ]; then
    OUTPUT_NAME="lessc-go.exe"
  fi

  echo "Building for $GOOS/$GOARCH..."

  GOOS=$GOOS GOARCH=$GOARCH go build \
    -ldflags="-s -w -X main.version=$VERSION" \
    -o "$OUTPUT_DIR/$SUFFIX/bin/$OUTPUT_NAME" \
    ./cmd/lessc-go

  # Copy runtime files to the bin directory
  # These are needed for JavaScript plugin support
  echo "Copying runtime files for $SUFFIX..."
  cp "$RUNTIME_DIR/plugin-host.js" "$OUTPUT_DIR/$SUFFIX/bin/"
  cp "$RUNTIME_DIR/bindings.js" "$OUTPUT_DIR/$SUFFIX/bin/"
done

echo "Done! Binaries and runtime files built for all platforms."
