# Dify Plugin Daemon

[English](README.md) | [中文](README_zh-CN.md)

Dify 平台的插件守护程序，用于运行自定义插件。

## 快速安装

### 方式一：使用安装脚本（推荐）

```bash
curl -fsSL https://raw.githubusercontent.com/langgenius/dify-plugin-daemon/main/install.sh | bash
```

或指定版本安装：

```bash
VERSION=0.0.1 curl -fsSL https://raw.githubusercontent.com/langgenius/dify-plugin-daemon/main/install.sh | bash
```

或使用 wget：

```bash
wget -qO- https://raw.githubusercontent.com/langgenius/dify-plugin-daemon/main/install.sh | bash
```

### 方式二：手动安装

1. 从[发布页面](https://github.com/langgenius/dify-plugin-daemon/releases)下载对应平台的二进制文件
2. 将其移动到相应目录并重命名为 `dify`：
   - macOS系统：`~/.local/bin/dify`
   - Linux系统：`/usr/local/bin/dify`
3. 添加可执行权限：`chmod +x dify`

## 安装详情

安装脚本会：

- 自动检测您的操作系统（macOS 或 Linux）和架构（AMD64 或 ARM64）
- 下载对应的二进制文件
- 安装到正确的位置：
  - macOS：`~/.local/bin/dify`
  - Linux：`/usr/local/bin/dify`
- 设置正确的权限
- 添加安装目录到 PATH（如果需要，仅限 macOS）

### 环境变量

- `VERSION`: 指定要安装的版本（默认：0.0.1）
  ```bash
  VERSION=0.0.2 ./install.sh
  ```

## 系统要求

- 操作系统：macOS 或 Linux
- 架构：AMD64（x86_64）或 ARM64（aarch64）
- 所需工具：`curl` 或 `wget`（用于安装） 