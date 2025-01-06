package plugin

import (
	"strings"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func TestRenderPythonToolTemplate(t *testing.T) {
	manifest := &plugin_entities.PluginDeclaration{
		PluginDeclarationWithoutAdvancedFields: plugin_entities.PluginDeclarationWithoutAdvancedFields{
			Name:   "test",
			Author: "test",
			Description: plugin_entities.I18nObject{
				EnUS: "test",
			},
		},
	}

	content, err := renderTemplate(PYTHON_TOOL_PY_TEMPLATE, manifest, []string{""})
	if err != nil {
		t.Errorf("failed to render template: %v", err)
	}

	if !strings.Contains(content, "TestTool") {
		t.Errorf("template content does not contain TestTool, snakeToCamel failed")
	}

	content, err = renderTemplate(PYTHON_TOOL_PROVIDER_TEMPLATE, manifest, []string{""})
	if err != nil {
		t.Errorf("failed to render template: %v", err)
	}

	if !strings.Contains(content, "test") {
		t.Errorf("template content does not contain TestTool, snakeToCamel failed")
	}
}
