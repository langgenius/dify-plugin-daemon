package testutils

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/packager"
)

const (
	defaultArchiveTimeout = 90 * time.Second
	archiveSizeLimit      = 300 << 20 // 300MiB
	packageSizeLimit      = 200 << 20 // 200MiB for generated difypkg
	maxAssetSize          = packageSizeLimit
)

// DownloadRepoArchive downloads a repository zipball (defaults to main branch).
// If ref is empty, "main" is used.
// Returns zip bytes and the actual ref (for logging).
func DownloadRepoArchive(ctx context.Context, owner, repo, ref string) ([]byte, string, error) {
	if owner == "" || repo == "" {
		return nil, "", errors.New("owner/repo is required")
	}
	if ref == "" {
		ref = "main"
	}

	ctx, cancel := context.WithTimeout(ctx, defaultArchiveTimeout)
	defer cancel()

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/zipball/%s", owner, repo, ref)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create archive request: %w", err)
	}

	token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
	if token == "" {
		token = strings.TrimSpace(os.Getenv("GITHUB_API_TOKEN"))
	}
	req.Header.Set("User-Agent", "dify-plugin-daemon-e2e-tests")
	req.Header.Set("Accept", "application/vnd.github+json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download archive: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, "", fmt.Errorf("unexpected archive status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, archiveSizeLimit+1))
	if err != nil {
		return nil, "", fmt.Errorf("failed to read archive: %w", err)
	}
	if int64(len(data)) > archiveSizeLimit {
		return nil, "", fmt.Errorf("archive exceeds size limit (%d bytes)", archiveSizeLimit)
	}

	return data, ref, nil
}

// ExtractPluginSource extracts only the specified subdirectory (e.g., models/ollama) from a zip.
// Returns the plugin root directory and a cleanup function.
func ExtractPluginSource(zipBytes []byte, subdir string) (string, func(), error) {
	subdir = strings.Trim(subdir, "/")
	if subdir == "" {
		return "", nil, errors.New("subdir is required")
	}

	reader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse zip: %w", err)
	}

	var rootPrefix string
	for _, f := range reader.File {
		name := path.Clean(f.Name)
		parts := strings.Split(name, "/")
		if len(parts) > 0 && parts[0] != "" {
			rootPrefix = parts[0]
			break
		}
	}
	if rootPrefix == "" {
		return "", nil, errors.New("cannot identify zip root directory")
	}

	targetPrefix := path.Join(rootPrefix, subdir)
	baseDir, err := os.MkdirTemp("", "dify-plugin-src-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	cleanup := func() {
		os.RemoveAll(baseDir)
	}

	for _, f := range reader.File {
		name := path.Clean(f.Name)
		if name == targetPrefix {
			continue
		}

		if !strings.HasPrefix(name, targetPrefix+"/") {
			continue
		}

		rel := strings.TrimPrefix(name, targetPrefix+"/")
		if rel == "" {
			continue
		}

		destPath := filepath.Join(baseDir, rel)
		if !strings.HasPrefix(destPath, baseDir) {
			cleanup()
			return "", nil, fmt.Errorf("path traversal detected: %s", destPath)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, 0o755); err != nil {
				cleanup()
				return "", nil, fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			cleanup()
			return "", nil, fmt.Errorf("failed to create parent directory: %w", err)
		}

		rc, err := f.Open()
		if err != nil {
			cleanup()
			return "", nil, fmt.Errorf("failed to open compressed file: %w", err)
		}

		dst, err := os.Create(destPath)
		if err != nil {
			rc.Close()
			cleanup()
			return "", nil, fmt.Errorf("failed to create file: %w", err)
		}

		if _, err := io.Copy(dst, rc); err != nil {
			rc.Close()
			dst.Close()
			cleanup()
			return "", nil, fmt.Errorf("failed to write file: %w", err)
		}

		rc.Close()
		dst.Close()
	}

	manifestPath := filepath.Join(baseDir, "manifest.yaml")
	if _, err := os.Stat(manifestPath); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("manifest.yaml not found: %w", err)
	}

	return baseDir, cleanup, nil
}

// PackPluginFromDir packs a directory into a difypkg, returns package bytes and manifest.
func PackPluginFromDir(dir string) ([]byte, plugin_entities.PluginDeclaration, error) {
	decoder, err := decoder.NewFSPluginDecoder(dir)
	if err != nil {
		return nil, plugin_entities.PluginDeclaration{}, fmt.Errorf("failed to create FS decoder: %w", err)
	}
	manifest, err := decoder.Manifest()
	if err != nil {
		return nil, plugin_entities.PluginDeclaration{}, fmt.Errorf("failed to read manifest: %w", err)
	}

	packager := packager.NewPackager(decoder)
	bundle, err := packager.Pack(packageSizeLimit)
	if err != nil {
		return nil, plugin_entities.PluginDeclaration{}, fmt.Errorf("failed to pack: %w", err)
	}

	return bundle, manifest, nil
}
