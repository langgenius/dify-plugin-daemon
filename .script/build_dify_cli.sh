#!/bin/bash
set -e

# Use environment variable VERSION if set, otherwise use default
VERSION=${VERSION:-v0.0.1}

# Output directory
OUTPUT_DIR=${OUTPUT_DIR:-dist}
mkdir -p "$OUTPUT_DIR"

# Platforms to build (OS/ARCH)
PLATFORMS=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
)

echo "Building dify-cli ${VERSION} for all platforms..."

for PLATFORM in "${PLATFORMS[@]}"; do
    GOOS=${PLATFORM%/*}
    GOARCH=${PLATFORM#*/}
    OUTPUT_NAME="dify-cli-${GOOS}-${GOARCH}"

    echo "Building ${OUTPUT_NAME}..."
    GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags "-s -w" \
        -o "${OUTPUT_DIR}/${OUTPUT_NAME}" ./cmd/dify_cli
    echo "✓ ${OUTPUT_NAME}"
done

echo ""
echo "Build complete:"
ls -lh "$OUTPUT_DIR"/dify-cli-*
