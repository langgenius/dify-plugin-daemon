package plugin

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/strings"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
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

//go:embed templates/python/model_provider.py
var PYTHON_MODEL_PROVIDER_PY_TEMPLATE []byte

//go:embed templates/python/model_provider.yaml
var PYTHON_MODEL_PROVIDER_TEMPLATE []byte

//go:embed templates/python/llm.py
var PYTHON_LLM_TEMPLATE []byte

//go:embed templates/python/llm.yaml
var PYTHON_LLM_MANIFEST_TEMPLATE []byte

//go:embed templates/python/text-embedding.py
var PYTHON_TEXT_EMBEDDING_TEMPLATE []byte

//go:embed templates/python/text-embedding.yaml
var PYTHON_TEXT_EMBEDDING_MANIFEST_TEMPLATE []byte

//go:embed templates/python/rerank.py
var PYTHON_RERANK_TEMPLATE []byte

//go:embed templates/python/rerank.yaml
var PYTHON_RERANK_MANIFEST_TEMPLATE []byte

//go:embed templates/python/tts.py
var PYTHON_TTS_TEMPLATE []byte

//go:embed templates/python/tts.yaml
var PYTHON_TTS_MANIFEST_TEMPLATE []byte

//go:embed templates/python/speech2text.py
var PYTHON_SPEECH2TEXT_TEMPLATE []byte

//go:embed templates/python/speech2text.yaml
var PYTHON_SPEECH2TEXT_MANIFEST_TEMPLATE []byte

//go:embed templates/python/moderation.py
var PYTHON_MODERATION_TEMPLATE []byte

//go:embed templates/python/moderation.yaml
var PYTHON_MODERATION_MANIFEST_TEMPLATE []byte

//go:embed templates/python/endpoint_group.yaml
var PYTHON_ENDPOINT_GROUP_MANIFEST_TEMPLATE []byte

//go:embed templates/python/endpoint.py
var PYTHON_ENDPOINT_TEMPLATE []byte

//go:embed templates/python/endpoint.yaml
var PYTHON_ENDPOINT_MANIFEST_TEMPLATE []byte

//go:embed templates/python/agent_provider.yaml
var PYTHON_AGENT_PROVIDER_MANIFEST_TEMPLATE []byte

//go:embed templates/python/agent_strategy.yaml
var PYTHON_AGENT_STRATEGY_MANIFEST_TEMPLATE []byte

//go:embed templates/python/agent_strategy.py
var PYTHON_AGENT_STRATEGY_TEMPLATE []byte

//go:embed templates/python/datasource.yaml
var PYTHON_DATASOURCE_MANIFEST_TEMPLATE []byte

//go:embed templates/python/datasource.py
var PYTHON_DATASOURCE_TEMPLATE []byte

//go:embed templates/python/datasource_provider.yaml
var PYTHON_DATASOURCE_PROVIDER_MANIFEST_TEMPLATE []byte

//go:embed templates/python/datasource_provider.py
var PYTHON_DATASOURCE_PROVIDER_PY_TEMPLATE []byte

//go:embed templates/python/GUIDE.md
var PYTHON_GUIDE []byte

//go:embed templates/python/.difyignore
var PYTHON_DIFYIGNORE []byte

//go:embed templates/python/.gitignore
var PYTHON_GITIGNORE []byte

func renderTemplate(
	original_template []byte, manifest *plugin_entities.PluginDeclaration, supported_model_types []string,
) (string, error) {
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"SnakeToCamel": parser.SnakeToCamel,
		"HasSubstring": func(substring string, haystack []string) bool {
			return strings.Find(haystack, substring)
		},
	}).Parse(string(original_template)))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]interface{}{
		"PluginName":          manifest.Name,
		"Author":              manifest.Author,
		"PluginDescription":   manifest.Description.EnUS,
		"SupportedModelTypes": supported_model_types,
		"Version":             manifest.Version,
		"Category":            manifest.Category(),
	}); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func writeFile(path string, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func createPythonEnvironment(
	root string, entrypoint string, manifest *plugin_entities.PluginDeclaration, category string,
) error {
	guide, err := renderTemplate(PYTHON_GUIDE, manifest, []string{})
	if err != nil {
		return err
	}

	if err := writeFile(filepath.Join(root, "GUIDE.md"), guide); err != nil {
		return err
	}

	// create the python environment
	entrypointFilePath := filepath.Join(root, fmt.Sprintf("%s.py", entrypoint))
	if err := os.WriteFile(entrypointFilePath, PYTHON_ENTRYPOINT_TEMPLATE, 0o644); err != nil {
		return err
	}

	requirementsFilePath := filepath.Join(root, "requirements.txt")
	if err := os.WriteFile(requirementsFilePath, PYTHON_REQUIREMENTS_TEMPLATE, 0o644); err != nil {
		return err
	}

	if err := writeFile(filepath.Join(root, ".difyignore"), string(PYTHON_DIFYIGNORE)); err != nil {
		return err
	}

	if err := writeFile(filepath.Join(root, ".gitignore"), string(PYTHON_GITIGNORE)); err != nil {
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

	if category == "extension" {
		if err := createPythonEndpointGroup(root, manifest); err != nil {
			return err
		}

		if err := createPythonEndpoint(root, manifest); err != nil {
			return err
		}
	}

	if category == "llm" || category == "text-embedding" || category == "speech2text" || category == "moderation" || category == "rerank" || category == "tts" {
		if err := createPythonModelProvider(root, manifest, []string{category}); err != nil {
			return err
		}
	}

	if category == "llm" {
		if err := createPythonLLM(root, manifest); err != nil {
			return err
		}
	}

	if category == "text-embedding" {
		if err := createPythonTextEmbedding(root, manifest); err != nil {
			return err
		}
	}

	if category == "speech2text" {
		if err := createPythonSpeech2Text(root, manifest); err != nil {
			return err
		}
	}

	if category == "moderation" {
		if err := createPythonModeration(root, manifest); err != nil {
			return err
		}
	}

	if category == "rerank" {
		if err := createPythonRerank(root, manifest); err != nil {
			return err
		}
	}

	if category == "tts" {
		if err := createPythonTTS(root, manifest); err != nil {
			return err
		}
	}

	if category == "agent-strategy" {
		if err := createPythonAgentStrategy(root, manifest); err != nil {
			return err
		}
	}

	if category == "datasource" {
		if err := createPythonDatasource(root, manifest); err != nil {
			return err
		}
	}

	return nil
}
