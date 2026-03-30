package media_transport

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/langgenius/dify-cloud-kit/oss"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

type InstalledBucket struct {
	oss           oss.OSS
	installedPath string
}

func NewInstalledBucket(oss oss.OSS, installedPath string) *InstalledBucket {
	// throw a warning if installed_path starts with non-alphanumeric characters
	if len(installedPath) > 0 {
		firstChar := installedPath[0]
		if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(string(firstChar)) {
			log.Warn("installed_path starts with non-alphanumeric characters", "path", installedPath)
		}
	} else {
		log.Warn("installed_path is empty", "path", "")
	}
	return &InstalledBucket{oss: oss, installedPath: installedPath}
}

// Save saves the plugin to the installed bucket
func (b *InstalledBucket) Save(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	file []byte,
) error {
	return b.oss.Save(filepath.Join(b.installedPath, pluginUniqueIdentifier.String()), file)
}

// Exists checks if the plugin exists in the installed bucket
func (b *InstalledBucket) Exists(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (bool, error) {
	return b.oss.Exists(filepath.Join(b.installedPath, pluginUniqueIdentifier.String()))
}

// Delete deletes the plugin from the installed bucket
func (b *InstalledBucket) Delete(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) error {
	return b.oss.Delete(filepath.Join(b.installedPath, pluginUniqueIdentifier.String()))
}

// Get gets the plugin from the installed bucket
func (b *InstalledBucket) Get(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) ([]byte, error) {
	return b.oss.Load(filepath.Join(b.installedPath, pluginUniqueIdentifier.String()))
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
			log.Error("failed to create PluginUniqueIdentifier from path", "path", path.Path, "error", err)
			continue
		}
		identifiers = append(identifiers, identifier)
	}
	return identifiers, nil
}
