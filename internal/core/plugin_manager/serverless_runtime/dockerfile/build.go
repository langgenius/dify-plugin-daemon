package dockerfile

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/strings"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/constants"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func handleTemplate(configuration *plugin_entities.PluginDeclaration, templateFunc func(configuration *plugin_entities.PluginDeclaration) (string, error)) (string, error) {
	if templateFunc == nil {
		return "", fmt.Errorf("template function is nil, language: %s, version: %s", configuration.Meta.Runner.Language, configuration.Meta.Runner.Version)
	}
	return templateFunc(configuration)
}

// GenerateDockerfile generates a Dockerfile for the plugin
func GenerateDockerfile(configuration *plugin_entities.PluginDeclaration) (string, error) {
	if !strings.Find(configuration.Meta.Arch, constants.AMD64) {
		return "", fmt.Errorf("unsupported architecture: %s", configuration.Meta.Arch)
	}

	switch configuration.Meta.Runner.Language {
	case constants.Python:
		return handleTemplate(configuration, pythonTemplates[configuration.Meta.Runner.Version])
	}

	return "", fmt.Errorf("unsupported language: %s", configuration.Meta.Runner.Language)
}
