#!/bin/bash
set -x

# Use environment variable VERSION if set, otherwise use default
VERSION=${VERSION:-v0.0.1}

# Detect OS and architecture
UNAME_OS=$(uname -s | tr '[:upper:]' '[:lower:]')
UNAME_ARCH=$(uname -m)

# Map OS
case "$UNAME_OS" in
  linux*)   GOOS=linux ;;
  darwin*)  GOOS=darwin ;;
  msys*|mingw*|cygwin*|windows*) GOOS=windows ;;
  *)        GOOS="$UNAME_OS" ;;
esac

# Map ARCH
case "$UNAME_ARCH" in
  x86_64|amd64) GOARCH=amd64 ;;
  i386|i686)    GOARCH=386 ;;
  aarch64)      GOARCH=arm64 ;;
  armv7l)       GOARCH=arm ;;
  *)            GOARCH="$UNAME_ARCH" ;;
esac

go build \
    -ldflags "-s -w" \
    -o "dify-cli-${GOOS}-${GOARCH}" ./cmd/dify_cli
