package dockerfile

import (
	_ "embed"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

var (
	pythonTemplates = map[string]func(configuration *plugin_entities.PluginDeclaration) (string, error){
		"3.12": GeneratePython312Dockerfile,
	}
)

//go:embed python312.dockerfile
var python312DockerfileTmpl string

// GeneratePython312Dockerfile generates a dockerfile for python 3.12
func GeneratePython312Dockerfile(configuration *plugin_entities.PluginDeclaration) (string, error) {
	entrypoint := configuration.Meta.Runner.Entrypoint

	return strings.Replace(python312DockerfileTmpl, "{{entrypoint}}", entrypoint, -1), nil
}
