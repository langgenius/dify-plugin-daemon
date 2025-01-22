#!/bin/bash

# Use environment variable VERSION if set, otherwise use default
VERSION=${VERSION:-0.0.1}

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Convert architecture naming
if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
else
    echo "Unsupported architecture: $ARCH"
    exit 1
fi

# Only allow macOS and Linux
if [ "$OS" != "darwin" ] && [ "$OS" != "linux" ]; then
    echo "Unsupported operating system: $OS"
    exit 1
fi

# Define download URL and binary name
BINARY_NAME="dify-plugin-$OS-$ARCH"
DOWNLOAD_URL="https://github.com/langgenius/dify-plugin-daemon/releases/download/$VERSION/$BINARY_NAME"

# Set installation directory based on OS
if [ "$OS" = "darwin" ]; then
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
    NEED_SUDO=false
else
    INSTALL_DIR="/usr/local/bin"
    # Check if we have write permission to /usr/local/bin
    if [ -w "$INSTALL_DIR" ]; then
        NEED_SUDO=false
    else
        NEED_SUDO=true
    fi
fi

# Create temporary directory for download
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR" || exit 1

# Download the binary
echo "Downloading $BINARY_NAME..."
if command -v curl >/dev/null 2>&1; then
    curl -L -o "dify-plugin-daemon" "$DOWNLOAD_URL"
elif command -v wget >/dev/null 2>&1; then
    wget -O "dify-plugin-daemon" "$DOWNLOAD_URL"
else
    echo "Error: Neither curl nor wget is installed"
    rm -rf "$TMP_DIR"
    exit 1
fi

# Make binary executable
chmod +x "dify-plugin-daemon"

# Install the binary with the new name
if [ "$NEED_SUDO" = true ]; then
    echo "Installing to $INSTALL_DIR (requires sudo)..."
    sudo mv "dify-plugin-daemon" "$INSTALL_DIR/dify-plugin"
else
    echo "Installing to $INSTALL_DIR..."
    mv "dify-plugin-daemon" "$INSTALL_DIR/dify-plugin"
fi

# Clean up
rm -rf "$TMP_DIR"

# For macOS, ensure ~/.local/bin is in PATH
if [ "$OS" = "darwin" ]; then
    if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
        SHELL_CONFIG=""
        if [ -f "$HOME/.zshrc" ]; then
            SHELL_CONFIG="$HOME/.zshrc"
        elif [ -f "$HOME/.bashrc" ]; then
            SHELL_CONFIG="$HOME/.bashrc"
        elif [ -f "$HOME/.bash_profile" ]; then
            SHELL_CONFIG="$HOME/.bash_profile"
        fi

        if [ -n "$SHELL_CONFIG" ]; then
            echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$SHELL_CONFIG"
            echo "Added $INSTALL_DIR to PATH in $SHELL_CONFIG"
            echo "Please run: source $SHELL_CONFIG"
        else
            echo "Please add the following line to your shell configuration file:"
            echo "export PATH=\"\$PATH:$INSTALL_DIR\""
        fi
    fi
fi

echo "Installation completed! The dify plugin daemon has been installed to $INSTALL_DIR/dify-plugin" 