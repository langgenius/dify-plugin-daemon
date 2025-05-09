#!/bin/bash
set -x

# Use environment variable VERSION if set, otherwise use default
VERSION=${VERSION:-v0.0.1}

# Detect OS and architecture
OS="${OS:-$(uname -s | tr '[:upper:]' '[:lower:]')}"
ARCH="${ARCH:-$(uname -m)}"

GOOS=${OS} GOARCH=${ARCH} \
go build \
    -ldflags "\
    -X 'main.VersionX=${VERSION}' "\
    -o dify-plugin-cli-${OS}-${ARCH} ./cmd/commandline
