package packager

import (
	"os"
	"path"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

func (p *Packager) ScanProvider() error {
	return nil
}

func (p *Packager) fetchManifest() (*plugin_entities.PluginDeclaration, error) {

	file_path := path.Join(p.wp, p.manifest)
	file, err := os.ReadFile(file_path)
	if err != nil {
		return nil, err
	}

	return plugin_entities.UnmarshalPluginDeclarationFromYaml(file)
}
