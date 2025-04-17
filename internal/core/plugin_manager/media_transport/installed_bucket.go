package media_transport

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/internal/oss"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

type InstalledBucket struct {
	oss           oss.OSS
	installedPath string
}

func NewInstalledBucket(oss oss.OSS, installed_path string) *InstalledBucket {
	return &InstalledBucket{oss: oss, installedPath: installed_path}
}

// Save saves the plugin to the installed bucket
func (b *InstalledBucket) Save(
	plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
	file []byte,
) error {
	compat_plugin_unique_identifier := plugin_unique_identifier.String()
	if runtime.GOOS == "windows" {
		compat_plugin_unique_identifier = strings.ReplaceAll(plugin_unique_identifier.String(), ":", "$")
	}
	return b.oss.Save(filepath.Join(b.installedPath, compat_plugin_unique_identifier), file)
}

// Exists checks if the plugin exists in the installed bucket
func (b *InstalledBucket) Exists(
	plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
) (bool, error) {
	compat_plugin_unique_identifier := plugin_unique_identifier.String()
	if runtime.GOOS == "windows" {
		compat_plugin_unique_identifier = strings.ReplaceAll(plugin_unique_identifier.String(), ":", "$")
	}
	return b.oss.Exists(filepath.Join(b.installedPath, compat_plugin_unique_identifier))
}

// Delete deletes the plugin from the installed bucket
func (b *InstalledBucket) Delete(
	plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
) error {
	compat_plugin_unique_identifier := plugin_unique_identifier.String()
	if runtime.GOOS == "windows" {
		compat_plugin_unique_identifier = strings.ReplaceAll(plugin_unique_identifier.String(), ":", "$")
	}
	return b.oss.Delete(filepath.Join(b.installedPath, compat_plugin_unique_identifier))
}

// Get gets the plugin from the installed bucket
func (b *InstalledBucket) Get(
	plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
) ([]byte, error) {
	compat_plugin_unique_identifier := plugin_unique_identifier.String()
	if runtime.GOOS == "windows" {
		compat_plugin_unique_identifier = strings.ReplaceAll(plugin_unique_identifier.String(), ":", "$")
	}
	return b.oss.Load(filepath.Join(b.installedPath, compat_plugin_unique_identifier))
}

// List lists all the plugins in the installed bucket
func (b *InstalledBucket) List() ([]plugin_entities.PluginUniqueIdentifier, error) {
	paths, err := b.oss.List(b.installedPath)
	if err != nil {
		return nil, err
	}
	identifiers := make([]plugin_entities.PluginUniqueIdentifier, 0)
	for _, path := range paths {
		if path.IsDir {
			continue
		}
		// skip hidden files
		if strings.HasPrefix(path.Path, ".") {
			continue
		}
		// remove prefix
		identifier, err := plugin_entities.NewPluginUniqueIdentifier(
			strings.TrimPrefix(path.Path, b.installedPath),
		)
		if err != nil {
			return nil, err
		}
		identifiers = append(identifiers, identifier)
	}
	return identifiers, nil
}
