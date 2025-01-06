package dockerfile

import (
	"strings"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/constants"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func preparePluginDeclaration() *plugin_entities.PluginDeclaration {
	return &plugin_entities.PluginDeclaration{
		PluginDeclarationWithoutAdvancedFields: plugin_entities.PluginDeclarationWithoutAdvancedFields{
			Meta: plugin_entities.PluginMeta{
				Arch: []constants.Arch{
					constants.AMD64,
				},
				Runner: plugin_entities.PluginRunner{
					Language:   constants.Python,
					Version:    "3.12",
					Entrypoint: "main",
				},
			},
		},
	}
}

func TestGenerateDockerfile(t *testing.T) {
	pluginDeclaration := preparePluginDeclaration()
	dockerfile, err := GenerateDockerfile(pluginDeclaration)
	if err != nil {
		t.Fatalf("Error generating Dockerfile: %v", err)
	}

	if !strings.Contains(dockerfile, "main") || !strings.Contains(dockerfile, "3.12") {
		t.Logf("Generated Dockerfile: %s", dockerfile)
	}
}

func TestGenerateDockerfileWithInvalidPluginDeclaration(t *testing.T) {
	pluginDeclaration := &plugin_entities.PluginDeclaration{}
	_, err := GenerateDockerfile(pluginDeclaration)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
}
