package init

import (
	"fmt"
	"path/filepath"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

func createPythonTool(root string, manifest *plugin_entities.PluginDeclaration) error {
	tool_file_content, err := renderTemplate(PYTHON_TOOL_PY_TEMPLATE, manifest, []string{""})
	if err != nil {
		return err
	}
	tool_file_path := filepath.Join(root, "tools", fmt.Sprintf("%s.py", manifest.Name))
	if err := writeFile(tool_file_path, tool_file_content); err != nil {
		return err
	}

	tool_manifest_file_path := filepath.Join(root, "tools", fmt.Sprintf("%s.yaml", manifest.Name))
	tool_manifest_file_content, err := renderTemplate(PYTHON_TOOL_TEMPLATE, manifest, []string{""})
	if err != nil {
		return err
	}
	if err := writeFile(tool_manifest_file_path, tool_manifest_file_content); err != nil {
		return err
	}

	return nil
}

func createPythonToolProvider(root string, manifest *plugin_entities.PluginDeclaration) error {
	tool_provider_file_content, err := renderTemplate(PYTHON_TOOL_PROVIDER_PY_TEMPLATE, manifest, []string{""})
	if err != nil {
		return err
	}
	tool_provider_file_path := filepath.Join(root, "provider", fmt.Sprintf("%s.py", manifest.Name))
	if err := writeFile(tool_provider_file_path, tool_provider_file_content); err != nil {
		return err
	}

	tool_provider_manifest_file_content, err := renderTemplate(PYTHON_TOOL_PROVIDER_TEMPLATE, manifest, []string{""})
	if err != nil {
		return err
	}
	tool_provider_manifest_file_path := filepath.Join(root, "provider", fmt.Sprintf("%s.yaml", manifest.Name))
	if err := writeFile(tool_provider_manifest_file_path, tool_provider_manifest_file_content); err != nil {
		return err
	}

	return nil
}

func createPythonEndpointGroup(root string, manifest *plugin_entities.PluginDeclaration) error {
	endpoint_group_file_content, err := renderTemplate(PYTHON_ENDPOINT_GROUP_MANIFEST_TEMPLATE, manifest, []string{""})
	if err != nil {
		return err
	}
	endpoint_group_file_path := filepath.Join(root, "group", fmt.Sprintf("%s.yaml", manifest.Name))
	if err := writeFile(endpoint_group_file_path, endpoint_group_file_content); err != nil {
		return err
	}

	return nil
}

func createPythonEndpoint(root string, manifest *plugin_entities.PluginDeclaration) error {
	endpoint_file_content, err := renderTemplate(PYTHON_ENDPOINT_MANIFEST_TEMPLATE, manifest, []string{""})
	if err != nil {
		return err
	}
	endpoint_file_path := filepath.Join(root, "endpoints", fmt.Sprintf("%s.yaml", manifest.Name))
	if err := writeFile(endpoint_file_path, endpoint_file_content); err != nil {
		return err
	}

	endpoint_py_file_content, err := renderTemplate(PYTHON_ENDPOINT_TEMPLATE, manifest, []string{""})
	if err != nil {
		return err
	}
	endpoint_py_file_path := filepath.Join(root, "endpoints", fmt.Sprintf("%s.py", manifest.Name))
	if err := writeFile(endpoint_py_file_path, endpoint_py_file_content); err != nil {
		return err
	}

	return nil
}

func createPythonLLM(root string, manifest *plugin_entities.PluginDeclaration) error {
	llm_file_content, err := renderTemplate(PYTHON_LLM_MANIFEST_TEMPLATE, manifest, []string{"llm"})
	if err != nil {
		return err
	}
	llm_file_path := filepath.Join(root, "models", "llm", "llm.yaml")
	if err := writeFile(llm_file_path, llm_file_content); err != nil {
		return err
	}

	llm_py_file_content, err := renderTemplate(PYTHON_LLM_TEMPLATE, manifest, []string{"llm"})
	if err != nil {
		return err
	}
	llm_py_file_path := filepath.Join(root, "models", "llm", "llm.py")
	if err := writeFile(llm_py_file_path, llm_py_file_content); err != nil {
		return err
	}

	return nil
}

func createPythonTextEmbedding(root string, manifest *plugin_entities.PluginDeclaration) error {
	text_embedding_file_content, err := renderTemplate(PYTHON_TEXT_EMBEDDING_MANIFEST_TEMPLATE, manifest, []string{"text_embedding"})
	if err != nil {
		return err
	}
	text_embedding_file_path := filepath.Join(root, "models", "text_embedding", "text_embedding.yaml")
	if err := writeFile(text_embedding_file_path, text_embedding_file_content); err != nil {
		return err
	}

	text_embedding_py_file_content, err := renderTemplate(PYTHON_TEXT_EMBEDDING_TEMPLATE, manifest, []string{"text_embedding"})
	if err != nil {
		return err
	}
	text_embedding_py_file_path := filepath.Join(root, "models", "text_embedding", "text_embedding.py")
	if err := writeFile(text_embedding_py_file_path, text_embedding_py_file_content); err != nil {
		return err
	}

	return nil
}

func createPythonRerank(root string, manifest *plugin_entities.PluginDeclaration) error {
	rerank_file_content, err := renderTemplate(PYTHON_RERANK_MANIFEST_TEMPLATE, manifest, []string{"rerank"})
	if err != nil {
		return err
	}
	rerank_file_path := filepath.Join(root, "models", "rerank", "rerank.yaml")
	if err := writeFile(rerank_file_path, rerank_file_content); err != nil {
		return err
	}

	rerank_py_file_content, err := renderTemplate(PYTHON_RERANK_TEMPLATE, manifest, []string{"rerank"})
	if err != nil {
		return err
	}
	rerank_py_file_path := filepath.Join(root, "models", "rerank", "rerank.py")
	if err := writeFile(rerank_py_file_path, rerank_py_file_content); err != nil {
		return err
	}

	return nil
}

func createPythonTTS(root string, manifest *plugin_entities.PluginDeclaration) error {
	tts_file_content, err := renderTemplate(PYTHON_TTS_MANIFEST_TEMPLATE, manifest, []string{"tts"})
	if err != nil {
		return err
	}
	tts_file_path := filepath.Join(root, "models", "tts", "tts.yaml")
	if err := writeFile(tts_file_path, tts_file_content); err != nil {
		return err
	}

	tts_py_file_content, err := renderTemplate(PYTHON_TTS_TEMPLATE, manifest, []string{"tts"})
	if err != nil {
		return err
	}
	tts_py_file_path := filepath.Join(root, "models", "tts", "tts.py")
	if err := writeFile(tts_py_file_path, tts_py_file_content); err != nil {
		return err
	}

	return nil
}

func createPythonSpeech2Text(root string, manifest *plugin_entities.PluginDeclaration) error {
	speech2text_file_content, err := renderTemplate(PYTHON_SPEECH2TEXT_MANIFEST_TEMPLATE, manifest, []string{"speech2text"})
	if err != nil {
		return err
	}
	speech2text_file_path := filepath.Join(root, "models", "speech2text", "speech2text.yaml")
	if err := writeFile(speech2text_file_path, speech2text_file_content); err != nil {
		return err
	}

	speech2text_py_file_content, err := renderTemplate(PYTHON_SPEECH2TEXT_TEMPLATE, manifest, []string{"speech2text"})
	if err != nil {
		return err
	}
	speech2text_py_file_path := filepath.Join(root, "models", "speech2text", "speech2text.py")
	if err := writeFile(speech2text_py_file_path, speech2text_py_file_content); err != nil {
		return err
	}

	return nil
}

func createPythonModeration(root string, manifest *plugin_entities.PluginDeclaration) error {
	moderation_file_content, err := renderTemplate(PYTHON_MODERATION_MANIFEST_TEMPLATE, manifest, []string{"moderation"})
	if err != nil {
		return err
	}
	moderation_file_path := filepath.Join(root, "models", "moderation", "moderation.yaml")
	if err := writeFile(moderation_file_path, moderation_file_content); err != nil {
		return err
	}

	moderation_py_file_content, err := renderTemplate(PYTHON_MODERATION_TEMPLATE, manifest, []string{"moderation"})
	if err != nil {
		return err
	}
	moderation_py_file_path := filepath.Join(root, "models", "moderation", "moderation.py")
	if err := writeFile(moderation_py_file_path, moderation_py_file_content); err != nil {
		return err
	}

	return nil
}

func createPythonModelProvider(root string, manifest *plugin_entities.PluginDeclaration, supported_model_types []string) error {
	provider_file_content, err := renderTemplate(PYTHON_MODEL_PROVIDER_PY_TEMPLATE, manifest, supported_model_types)
	if err != nil {
		return err
	}
	provider_file_path := filepath.Join(root, "provider", fmt.Sprintf("%s.py", manifest.Name))
	if err := writeFile(provider_file_path, provider_file_content); err != nil {
		return err
	}

	provider_manifest_file_content, err := renderTemplate(PYTHON_MODEL_PROVIDER_TEMPLATE, manifest, supported_model_types)
	if err != nil {
		return err
	}
	provider_manifest_file_path := filepath.Join(root, "provider", fmt.Sprintf("%s.yaml", manifest.Name))
	if err := writeFile(provider_manifest_file_path, provider_manifest_file_content); err != nil {
		return err
	}

	return nil
}
