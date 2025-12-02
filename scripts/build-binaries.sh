#!/bin/bash
# Cross-compile lessc-go for all supported platforms

set -e

VERSION=${1:-"dev"}
OUTPUT_DIR="npm"

# Platform targets: GOOS/GOARCH -> npm package suffix
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
    -o "$OUTPUT_DIR/less.go-$SUFFIX/bin/$OUTPUT_NAME" \
    ./cmd/lessc-go
done

echo "Done! Binaries built for all platforms."
