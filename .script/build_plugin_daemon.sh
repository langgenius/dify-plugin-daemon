#!/bin/bash
set -x

# Use environment variable VERSION if set, otherwise use default
VERSION=${VERSION:-0.0.1}

# Detect OS and architecture
OS="${OS:-$(uname -s | tr '[:upper:]' '[:lower:]')}"
ARCH="${ARCH:-$(uname -m)}"

GOOS=${OS} GOARCH=${ARCH} \
go build \
    -ldflags "\
    -X 'github.com/langgenius/dify-plugin-daemon/internal/manifest.VersionX=${VERSION}' \
    -X 'github.com/langgenius/dify-plugin-daemon/internal/manifest.BuildTimeX=$(date -u +%Y-%m-%dT%H:%M:%S%z)'" \
    -o dify-plugin-daemon-${OS}-${ARCH} ./cmd/server/main.go
