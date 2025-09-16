package plugin

import (
	"fmt"
	"path/filepath"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func createPythonTool(root string, manifest *plugin_entities.PluginDeclaration) error {
	toolFileContent, err := renderTemplate(PYTHON_TOOL_PY_TEMPLATE, manifest, []string{""})
	if err != nil {
		return err
	}
	toolFilePath := filepath.Join(root, "tools", fmt.Sprintf("%s.py", manifest.Name))
	if err := writeFile(toolFilePath, toolFileContent); err != nil {
		return err
	}

	toolManifestFilePath := filepath.Join(root, "tools", fmt.Sprintf("%s.yaml", manifest.Name))
	toolManifestFileContent, err := renderTemplate(PYTHON_TOOL_TEMPLATE, manifest, []string{""})
	if err != nil {
		return err
	}
	if err := writeFile(toolManifestFilePath, toolManifestFileContent); err != nil {
		return err
	}

	return nil
}

func createPythonToolProvider(root string, manifest *plugin_entities.PluginDeclaration) error {
	toolProviderFileContent, err := renderTemplate(PYTHON_TOOL_PROVIDER_PY_TEMPLATE, manifest, []string{""})
	if err != nil {
		return err
	}
	toolProviderFilePath := filepath.Join(root, "provider", fmt.Sprintf("%s.py", manifest.Name))
	if err := writeFile(toolProviderFilePath, toolProviderFileContent); err != nil {
		return err
	}

	toolProviderManifestFileContent, err := renderTemplate(PYTHON_TOOL_PROVIDER_TEMPLATE, manifest, []string{""})
	if err != nil {
		return err
	}
	toolProviderManifestFilePath := filepath.Join(root, "provider", fmt.Sprintf("%s.yaml", manifest.Name))
	if err := writeFile(toolProviderManifestFilePath, toolProviderManifestFileContent); err != nil {
		return err
	}

	return nil
}

func createPythonEndpointGroup(root string, manifest *plugin_entities.PluginDeclaration) error {
	endpointGroupFileContent, err := renderTemplate(PYTHON_ENDPOINT_GROUP_MANIFEST_TEMPLATE, manifest, []string{""})
	if err != nil {
		return err
	}
	endpointGroupFilePath := filepath.Join(root, "group", fmt.Sprintf("%s.yaml", manifest.Name))
	if err := writeFile(endpointGroupFilePath, endpointGroupFileContent); err != nil {
		return err
	}

	return nil
}

func createPythonEndpoint(root string, manifest *plugin_entities.PluginDeclaration) error {
	endpointFileContent, err := renderTemplate(PYTHON_ENDPOINT_MANIFEST_TEMPLATE, manifest, []string{""})
	if err != nil {
		return err
	}
	endpointFilePath := filepath.Join(root, "endpoints", fmt.Sprintf("%s.yaml", manifest.Name))
	if err := writeFile(endpointFilePath, endpointFileContent); err != nil {
		return err
	}

	endpointPyFileContent, err := renderTemplate(PYTHON_ENDPOINT_TEMPLATE, manifest, []string{""})
	if err != nil {
		return err
	}
	endpointPyFilePath := filepath.Join(root, "endpoints", fmt.Sprintf("%s.py", manifest.Name))
	if err := writeFile(endpointPyFilePath, endpointPyFileContent); err != nil {
		return err
	}

	return nil
}

func createPythonLLM(root string, manifest *plugin_entities.PluginDeclaration) error {
	llmFileContent, err := renderTemplate(PYTHON_LLM_MANIFEST_TEMPLATE, manifest, []string{"llm"})
	if err != nil {
		return err
	}
	llmFilePath := filepath.Join(root, "models", "llm", "llm.yaml")
	if err := writeFile(llmFilePath, llmFileContent); err != nil {
		return err
	}

	llmPyFileContent, err := renderTemplate(PYTHON_LLM_TEMPLATE, manifest, []string{"llm"})
	if err != nil {
		return err
	}
	llmPyFilePath := filepath.Join(root, "models", "llm", "llm.py")
	if err := writeFile(llmPyFilePath, llmPyFileContent); err != nil {
		return err
	}

	return nil
}

func createPythonTextEmbedding(root string, manifest *plugin_entities.PluginDeclaration) error {
	textEmbeddingFileContent, err := renderTemplate(PYTHON_TEXT_EMBEDDING_MANIFEST_TEMPLATE, manifest, []string{"text_embedding"})
	if err != nil {
		return err
	}
	textEmbeddingFilePath := filepath.Join(root, "models", "text_embedding", "text_embedding.yaml")
	if err := writeFile(textEmbeddingFilePath, textEmbeddingFileContent); err != nil {
		return err
	}

	textEmbeddingPyFileContent, err := renderTemplate(PYTHON_TEXT_EMBEDDING_TEMPLATE, manifest, []string{"text_embedding"})
	if err != nil {
		return err
	}
	textEmbeddingPyFilePath := filepath.Join(root, "models", "text_embedding", "text_embedding.py")
	if err := writeFile(textEmbeddingPyFilePath, textEmbeddingPyFileContent); err != nil {
		return err
	}

	return nil
}

func createPythonRerank(root string, manifest *plugin_entities.PluginDeclaration) error {
	rerankFileContent, err := renderTemplate(PYTHON_RERANK_MANIFEST_TEMPLATE, manifest, []string{"rerank"})
	if err != nil {
		return err
	}
	rerankFilePath := filepath.Join(root, "models", "rerank", "rerank.yaml")
	if err := writeFile(rerankFilePath, rerankFileContent); err != nil {
		return err
	}

	rerankPyFileContent, err := renderTemplate(PYTHON_RERANK_TEMPLATE, manifest, []string{"rerank"})
	if err != nil {
		return err
	}
	rerankPyFilePath := filepath.Join(root, "models", "rerank", "rerank.py")
	if err := writeFile(rerankPyFilePath, rerankPyFileContent); err != nil {
		return err
	}

	return nil
}

func createPythonTTS(root string, manifest *plugin_entities.PluginDeclaration) error {
	ttsFileContent, err := renderTemplate(PYTHON_TTS_MANIFEST_TEMPLATE, manifest, []string{"tts"})
	if err != nil {
		return err
	}
	ttsFilePath := filepath.Join(root, "models", "tts", "tts.yaml")
	if err := writeFile(ttsFilePath, ttsFileContent); err != nil {
		return err
	}

	ttsPyFileContent, err := renderTemplate(PYTHON_TTS_TEMPLATE, manifest, []string{"tts"})
	if err != nil {
		return err
	}
	ttsPyFilePath := filepath.Join(root, "models", "tts", "tts.py")
	if err := writeFile(ttsPyFilePath, ttsPyFileContent); err != nil {
		return err
	}

	return nil
}

func createPythonSpeech2Text(root string, manifest *plugin_entities.PluginDeclaration) error {
	speech2textFileContent, err := renderTemplate(PYTHON_SPEECH2TEXT_MANIFEST_TEMPLATE, manifest, []string{"speech2text"})
	if err != nil {
		return err
	}
	speech2textFilePath := filepath.Join(root, "models", "speech2text", "speech2text.yaml")
	if err := writeFile(speech2textFilePath, speech2textFileContent); err != nil {
		return err
	}

	speech2textPyFileContent, err := renderTemplate(PYTHON_SPEECH2TEXT_TEMPLATE, manifest, []string{"speech2text"})
	if err != nil {
		return err
	}
	speech2textPyFilePath := filepath.Join(root, "models", "speech2text", "speech2text.py")
	if err := writeFile(speech2textPyFilePath, speech2textPyFileContent); err != nil {
		return err
	}

	return nil
}

func createPythonModeration(root string, manifest *plugin_entities.PluginDeclaration) error {
	moderationFileContent, err := renderTemplate(PYTHON_MODERATION_MANIFEST_TEMPLATE, manifest, []string{"moderation"})
	if err != nil {
		return err
	}
	moderationFilePath := filepath.Join(root, "models", "moderation", "moderation.yaml")
	if err := writeFile(moderationFilePath, moderationFileContent); err != nil {
		return err
	}

	moderationPyFileContent, err := renderTemplate(PYTHON_MODERATION_TEMPLATE, manifest, []string{"moderation"})
	if err != nil {
		return err
	}
	moderationPyFilePath := filepath.Join(root, "models", "moderation", "moderation.py")
	if err := writeFile(moderationPyFilePath, moderationPyFileContent); err != nil {
		return err
	}

	return nil
}

func createPythonModelProvider(root string, manifest *plugin_entities.PluginDeclaration, supported_model_types []string) error {
	providerFileContent, err := renderTemplate(PYTHON_MODEL_PROVIDER_PY_TEMPLATE, manifest, supported_model_types)
	if err != nil {
		return err
	}
	providerFilePath := filepath.Join(root, "provider", fmt.Sprintf("%s.py", manifest.Name))
	if err := writeFile(providerFilePath, providerFileContent); err != nil {
		return err
	}

	providerManifestFileContent, err := renderTemplate(PYTHON_MODEL_PROVIDER_TEMPLATE, manifest, supported_model_types)
	if err != nil {
		return err
	}
	providerManifestFilePath := filepath.Join(root, "provider", fmt.Sprintf("%s.yaml", manifest.Name))
	if err := writeFile(providerManifestFilePath, providerManifestFileContent); err != nil {
		return err
	}

	return nil
}

func createPythonAgentStrategy(root string, manifest *plugin_entities.PluginDeclaration) error {
	agentFileContent, err := renderTemplate(PYTHON_AGENT_PROVIDER_MANIFEST_TEMPLATE, manifest, []string{"agent"})
	if err != nil {
		return err
	}
	agentFilePath := filepath.Join(root, "provider", fmt.Sprintf("%s.yaml", manifest.Name))
	if err := writeFile(agentFilePath, agentFileContent); err != nil {
		return err
	}

	agentStrategyFileContent, err := renderTemplate(PYTHON_AGENT_STRATEGY_MANIFEST_TEMPLATE, manifest, []string{"agent"})
	if err != nil {
		return err
	}
	agentStrategyFilePath := filepath.Join(root, "strategies", fmt.Sprintf("%s.yaml", manifest.Name))
	if err := writeFile(agentStrategyFilePath, agentStrategyFileContent); err != nil {
		return err
	}

	agentStrategyPyFileContent, err := renderTemplate(PYTHON_AGENT_STRATEGY_TEMPLATE, manifest, []string{"agent"})
	if err != nil {
		return err
	}
	agentStrategyPyFilePath := filepath.Join(root, "strategies", fmt.Sprintf("%s.py", manifest.Name))
	if err := writeFile(agentStrategyPyFilePath, agentStrategyPyFileContent); err != nil {
		return err
	}

	return nil
}

func createPythonDatasource(root string, manifest *plugin_entities.PluginDeclaration) error {
	datasourceFileContent, err := renderTemplate(PYTHON_DATASOURCE_TEMPLATE, manifest, []string{""})
	if err != nil {
		return err
	}
	datasourceFilePath := filepath.Join(root, "datasources", fmt.Sprintf("%s.py", manifest.Name))
	if err := writeFile(datasourceFilePath, datasourceFileContent); err != nil {
		return err
	}

	datasourceManifestFilePath := filepath.Join(root, "datasources", fmt.Sprintf("%s.yaml", manifest.Name))
	datasourceManifestFileContent, err := renderTemplate(PYTHON_DATASOURCE_MANIFEST_TEMPLATE, manifest, []string{""})
	if err != nil {
		return err
	}
	if err := writeFile(datasourceManifestFilePath, datasourceManifestFileContent); err != nil {
		return err
	}

	return nil
}
