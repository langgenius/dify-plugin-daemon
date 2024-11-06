package init

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

//go:embed templates/python/main.py
var PYTHON_ENTRYPOINT_TEMPLATE []byte

//go:embed templates/python/requirements.txt
var PYTHON_REQUIREMENTS_TEMPLATE []byte

//go:embed templates/python/tool_provider.yaml
var PYTHON_TOOL_PROVIDER_TEMPLATE []byte

//go:embed templates/python/tool.yaml
var PYTHON_TOOL_TEMPLATE []byte

//go:embed templates/python/tool.py
var PYTHON_TOOL_PY_TEMPLATE []byte

//go:embed templates/python/tool_provider.py
var PYTHON_TOOL_PROVIDER_PY_TEMPLATE []byte

func createPythonEnvironment(
	root string, entrypoint string, manifest *plugin_entities.PluginDeclaration, category string,
) error {
	// TODO: enhance to use template renderer

	// create the python environment
	entrypoint_file_path := filepath.Join(root, fmt.Sprintf("%s.py", entrypoint))
	if err := os.WriteFile(entrypoint_file_path, PYTHON_ENTRYPOINT_TEMPLATE, 0o644); err != nil {
		return err
	}

	requirements_file_path := filepath.Join(root, "requirements.txt")
	if err := os.WriteFile(requirements_file_path, PYTHON_REQUIREMENTS_TEMPLATE, 0o644); err != nil {
		return err
	}

	if category == "tool" {
		if err := createPythonTool(root, manifest); err != nil {
			return err
		}

		if err := createPythonToolProvider(root, manifest); err != nil {
			return err
		}
	}

	return nil
}

func createPythonTool(root string, manifest *plugin_entities.PluginDeclaration) error {
	// create the tool
	tool_dir := filepath.Join(root, "tools")
	if err := os.MkdirAll(tool_dir, 0o755); err != nil {
		return err
	}
	// replace the plugin name/author/description in the template
	tool_file_content := strings.ReplaceAll(
		string(PYTHON_TOOL_PY_TEMPLATE), "{{plugin_name}}", manifest.Name,
	)
	tool_file_content = strings.ReplaceAll(
		tool_file_content, "{{author}}", manifest.Author,
	)
	tool_file_content = strings.ReplaceAll(
		tool_file_content, "{{plugin_description}}", manifest.Description.EnUS,
	)
	tool_file_path := filepath.Join(tool_dir, fmt.Sprintf("%s.py", manifest.Name))
	if err := os.WriteFile(tool_file_path, []byte(tool_file_content), 0o644); err != nil {
		return err
	}

	// create the tool manifest
	tool_manifest_file_path := filepath.Join(tool_dir, fmt.Sprintf("%s.yaml", manifest.Name))
	if err := os.WriteFile(tool_manifest_file_path, PYTHON_TOOL_TEMPLATE, 0o644); err != nil {
		return err
	}
	tool_manifest_file_content := strings.ReplaceAll(
		string(PYTHON_TOOL_TEMPLATE), "{{plugin_name}}", manifest.Name,
	)
	tool_manifest_file_content = strings.ReplaceAll(
		tool_manifest_file_content, "{{author}}", manifest.Author,
	)
	tool_manifest_file_content = strings.ReplaceAll(
		tool_manifest_file_content, "{{plugin_description}}", manifest.Description.EnUS,
	)
	if err := os.WriteFile(tool_manifest_file_path, []byte(tool_manifest_file_content), 0o644); err != nil {
		return err
	}

	return nil
}

func createPythonToolProvider(root string, manifest *plugin_entities.PluginDeclaration) error {
	// create the tool provider
	tool_provider_dir := filepath.Join(root, "provider")
	if err := os.MkdirAll(tool_provider_dir, 0o755); err != nil {
		return err
	}
	// replace the plugin name/author/description in the template
	tool_provider_file_content := strings.ReplaceAll(
		string(PYTHON_TOOL_PROVIDER_PY_TEMPLATE), "{{plugin_name}}", manifest.Name,
	)
	tool_provider_file_content = strings.ReplaceAll(
		tool_provider_file_content, "{{author}}", manifest.Author,
	)
	tool_provider_file_content = strings.ReplaceAll(
		tool_provider_file_content, "{{plugin_description}}", manifest.Description.EnUS,
	)
	tool_provider_file_path := filepath.Join(tool_provider_dir, fmt.Sprintf("%s.py", manifest.Name))
	if err := os.WriteFile(tool_provider_file_path, []byte(tool_provider_file_content), 0o644); err != nil {
		return err
	}

	// create the tool provider manifest
	tool_provider_manifest_file_path := filepath.Join(tool_provider_dir, fmt.Sprintf("%s.yaml", manifest.Name))
	if err := os.WriteFile(tool_provider_manifest_file_path, PYTHON_TOOL_PROVIDER_TEMPLATE, 0o644); err != nil {
		return err
	}
	tool_provider_manifest_file_content := strings.ReplaceAll(
		string(PYTHON_TOOL_PROVIDER_TEMPLATE), "{{plugin_name}}", manifest.Name,
	)
	tool_provider_manifest_file_content = strings.ReplaceAll(
		tool_provider_manifest_file_content, "{{author}}", manifest.Author,
	)
	tool_provider_manifest_file_content = strings.ReplaceAll(
		tool_provider_manifest_file_content, "{{plugin_description}}", manifest.Description.EnUS,
	)
	if err := os.WriteFile(tool_provider_manifest_file_path, []byte(tool_provider_manifest_file_content), 0o644); err != nil {
		return err
	}

	return nil
}
