package packager

import (
	"path"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func (p *Packager) ScanProvider() error {
	return nil
}

func (p *Packager) fetchManifest() (*plugin_entities.PluginDeclaration, error) {
	file, err := p.decoder.ReadFile(path.Clean(p.manifest))
	if err != nil {
		return nil, err
	}

	return plugin_entities.UnmarshalPluginDeclarationFromYaml(file)
}
