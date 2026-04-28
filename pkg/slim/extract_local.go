package slim

import (
	"errors"
	"fmt"
	"os"

	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

func ExtractLocal(opts ExtractOptions, local *LocalConfig) (*ExtractResult, error) {
	if opts.PluginID != "" && opts.Path != "" {
		return nil, NewError(ErrInvalidInput, "only one of -id or -path may be provided")
	}
	if opts.PluginID == "" && opts.Path == "" {
		return nil, NewError(ErrInvalidInput, "one of -id or -path is required")
	}

	if opts.Path != "" {
		return extractLocalPath(opts.Path, "")
	}
	if local.Folder == "" {
		return nil, NewError(ErrConfigInvalid, "local.folder is required when extract uses -id")
	}

	return extractLocalPath(pluginWorkingPath(local.Folder, opts.PluginID), opts.PluginID)
}

func extractLocalPath(path string, pluginID string) (*ExtractResult, error) {
	stat, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, NewError(ErrPluginNotFound, fmt.Sprintf("plugin path not found: %s", path))
		}
		return nil, NewError(ErrInvalidInput, fmt.Sprintf("stat plugin path: %s", err))
	}

	if stat.IsDir() {
		dec, err := decoder.NewFSPluginDecoder(path)
		if err != nil {
			return nil, NewError(ErrPluginPackageInvalid, fmt.Sprintf("decode plugin directory: %s", err))
		}
		return extractFromDecoder(dec, pluginID)
	}

	pkgBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, NewError(ErrInvalidInput, fmt.Sprintf("read plugin package: %s", err))
	}

	dec, err := decoder.NewZipPluginDecoder(pkgBytes)
	if err != nil {
		return nil, NewError(ErrPluginPackageInvalid, fmt.Sprintf("decode plugin package: %s", err))
	}
	defer dec.Close()

	return extractFromDecoder(dec, pluginID)
}

func extractFromDecoder(dec decoder.PluginDecoder, pluginID string) (*ExtractResult, error) {
	manifest, err := dec.Manifest()
	if err != nil {
		return nil, NewError(ErrPluginPackageInvalid, fmt.Sprintf("read plugin manifest: %s", err))
	}

	result := &ExtractResult{
		UniqueIdentifier: pluginID,
		Manifest:         manifest,
	}

	if result.UniqueIdentifier == "" {
		identity, err := dec.UniqueIdentity()
		if err == nil {
			result.UniqueIdentifier = identity.String()
		}
	}

	verification, _ := dec.Verification()
	if verification == nil && dec.Verified() {
		verification = decoder.DefaultVerification()
	}
	result.Verification = verification

	return result, nil
}
