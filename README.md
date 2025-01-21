# Dify Plugin Daemon

[English](README.md) | [中文](README_zh-CN.md)

A plugin daemon for Dify platform that enables running custom plugins.

## Quick Installation

### Option 1: Using Install Script (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/langgenius/dify-plugin-daemon/main/install.sh | bash
```

or with specific version:

```bash
VERSION=0.0.1 curl -fsSL https://raw.githubusercontent.com/langgenius/dify-plugin-daemon/main/install.sh | bash
```

or using wget:

```bash
wget -qO- https://raw.githubusercontent.com/langgenius/dify-plugin-daemon/main/install.sh | bash
```

### Option 2: Manual Installation

1. Download the binary for your platform from [releases page](https://github.com/langgenius/dify-plugin-daemon/releases)
2. Move it to the appropriate directory and rename it to `dify`:
   - For macOS: `~/.local/bin/dify`
   - For Linux: `/usr/local/bin/dify`
3. Make it executable: `chmod +x dify`

## Installation Details

The installation script will:

- Automatically detect your operating system (macOS or Linux) and architecture (AMD64 or ARM64)
- Download the appropriate binary
- Install it to the correct location:
  - macOS: `~/.local/bin/dify`
  - Linux: `/usr/local/bin/dify`
- Set the correct permissions
- Add the installation directory to PATH (if needed, for macOS)

### Environment Variables

- `VERSION`: Specify the version to install (default: 0.0.1)
  ```bash
  VERSION=0.0.2 ./install.sh
  ```

## System Requirements

- Operating System: macOS or Linux
- Architecture: AMD64 (x86_64) or ARM64 (aarch64)
- Required tools: `curl` or `wget` (for installation)
