package media_transport

import (
	"path/filepath"
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
	return b.oss.Save(filepath.Join(b.installedPath, plugin_unique_identifier.String()), file)
}

// Exists checks if the plugin exists in the installed bucket
func (b *InstalledBucket) Exists(
	plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
) (bool, error) {
	return b.oss.Exists(filepath.Join(b.installedPath, plugin_unique_identifier.String()))
}

// Delete deletes the plugin from the installed bucket
func (b *InstalledBucket) Delete(
	plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
) error {
	return b.oss.Delete(filepath.Join(b.installedPath, plugin_unique_identifier.String()))
}

// Get gets the plugin from the installed bucket
func (b *InstalledBucket) Get(
	plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
) ([]byte, error) {
	return b.oss.Load(filepath.Join(b.installedPath, plugin_unique_identifier.String()))
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
