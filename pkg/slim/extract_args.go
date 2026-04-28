package slim

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func BuildActionArgs(manifest plugin_entities.PluginDeclaration, action string) (map[string]any, error) {
	switch action {
	case "invoke_tool":
		_, tool, err := firstTool(manifest)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"provider":        manifest.Tool.Identity.Name,
			"tool":            tool.Identity.Name,
			"tool_parameters": toolParameters(tool.Parameters),
			"credentials":     map[string]any{},
		}, nil
	case "validate_tool_credentials":
		if manifest.Tool == nil {
			return nil, NewError(ErrInvalidInput, "plugin has no tool provider")
		}
		return map[string]any{
			"provider":    manifest.Tool.Identity.Name,
			"credentials": map[string]any{},
		}, nil
	case "get_tool_runtime_parameters":
		_, tool, err := firstTool(manifest)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"provider":    manifest.Tool.Identity.Name,
			"tool":        tool.Identity.Name,
			"credentials": map[string]any{},
		}, nil
	case "get_ai_model_schemas", "invoke_llm", "get_llm_num_tokens",
		"invoke_text_embedding", "invoke_multimodal_embedding", "get_text_embedding_num_tokens",
		"invoke_rerank", "invoke_multimodal_rerank", "invoke_tts", "get_tts_model_voices",
		"invoke_speech2text", "invoke_moderation", "validate_provider_credentials",
		"validate_model_credentials":
		return modelActionArgs(manifest, action)
	case "invoke_agent_strategy":
		if manifest.AgentStrategy == nil || len(manifest.AgentStrategy.Strategies) == 0 {
			return nil, NewError(ErrInvalidInput, "plugin has no agent strategy")
		}
		strategy := manifest.AgentStrategy.Strategies[0]
		return map[string]any{
			"agent_strategy_provider": manifest.AgentStrategy.Identity.Name,
			"agent_strategy":          strategy.Identity.Name,
			"agent_strategy_params":   agentStrategyParameters(strategy.Parameters),
		}, nil
	default:
		return nil, NewError(ErrUnknownAction, action)
	}
}

func firstTool(manifest plugin_entities.PluginDeclaration) (*plugin_entities.ToolProviderDeclaration, plugin_entities.ToolDeclaration, error) {
	if manifest.Tool == nil {
		return nil, plugin_entities.ToolDeclaration{}, NewError(ErrInvalidInput, "plugin has no tool provider")
	}
	if len(manifest.Tool.Tools) == 0 {
		return nil, plugin_entities.ToolDeclaration{}, NewError(ErrInvalidInput, "tool provider has no tools")
	}
	return manifest.Tool, manifest.Tool.Tools[0], nil
}

func modelActionArgs(manifest plugin_entities.PluginDeclaration, action string) (map[string]any, error) {
	if manifest.Model == nil {
		return nil, NewError(ErrInvalidInput, "plugin has no model provider")
	}

	modelType := modelTypeForAction(action)
	model, err := firstModel(manifest.Model, modelType)
	if err != nil {
		return nil, err
	}

	args := map[string]any{
		"provider":    manifest.Model.Provider,
		"model":       model.Model,
		"model_type":  string(model.ModelType),
		"credentials": map[string]any{},
	}

	switch action {
	case "validate_provider_credentials":
		delete(args, "model")
		delete(args, "model_type")
	case "invoke_llm":
		args["model_parameters"] = map[string]any{}
		args["prompt_messages"] = []map[string]any{
			{"role": "user", "content": "hello"},
		}
		args["tools"] = []any{}
		args["stop"] = []string{}
		args["stream"] = false
	case "get_llm_num_tokens":
		args["prompt_messages"] = []map[string]any{
			{"role": "user", "content": "hello"},
		}
		args["tools"] = []any{}
	case "invoke_text_embedding", "get_text_embedding_num_tokens":
		args["texts"] = []string{"hello"}
		if action == "invoke_text_embedding" {
			args["input_type"] = "query"
		}
	case "invoke_multimodal_embedding":
		args["documents"] = []map[string]any{{"type": "text", "text": "hello"}}
		args["input_type"] = "query"
	case "invoke_rerank":
		args["query"] = "hello"
		args["docs"] = []string{"hello"}
	case "invoke_multimodal_rerank":
		args["query"] = map[string]any{"type": "text", "text": "hello"}
		args["docs"] = []map[string]any{{"type": "text", "text": "hello"}}
	case "invoke_tts":
		args["content_text"] = "hello"
		args["voice"] = "alloy"
		args["tenant_id"] = "00000000-0000-0000-0000-000000000000"
	case "invoke_speech2text":
		args["file"] = ""
	case "invoke_moderation":
		args["text"] = "hello"
	}

	return args, nil
}

func modelTypeForAction(action string) plugin_entities.ModelType {
	switch action {
	case "invoke_text_embedding", "get_text_embedding_num_tokens":
		return plugin_entities.MODEL_TYPE_TEXT_EMBEDDING
	case "invoke_multimodal_embedding":
		return plugin_entities.MODEL_TYPE_MULTIMODAL_EMBEDDING
	case "invoke_rerank":
		return plugin_entities.MODEL_TYPE_RERANKING
	case "invoke_multimodal_rerank":
		return plugin_entities.MODEL_TYPE_MULTIMODAL_RERANK
	case "invoke_tts", "get_tts_model_voices":
		return plugin_entities.MODEL_TYPE_TTS
	case "invoke_speech2text":
		return plugin_entities.MODEL_TYPE_SPEECH2TEXT
	case "invoke_moderation":
		return plugin_entities.MODEL_TYPE_MODERATION
	default:
		return plugin_entities.MODEL_TYPE_LLM
	}
}

func firstModel(provider *plugin_entities.ModelProviderDeclaration, modelType plugin_entities.ModelType) (plugin_entities.ModelDeclaration, error) {
	for _, model := range provider.Models {
		if model.ModelType == modelType {
			return model, nil
		}
	}
	if len(provider.Models) > 0 {
		return provider.Models[0], nil
	}
	return plugin_entities.ModelDeclaration{}, NewError(ErrInvalidInput, fmt.Sprintf("model provider has no %s models", modelType))
}

func toolParameters(params []plugin_entities.ToolParameter) map[string]any {
	out := map[string]any{}
	for _, param := range params {
		out[param.Name] = defaultValue(param.Default, string(param.Type), param.Options)
	}
	return out
}

func agentStrategyParameters(params []plugin_entities.AgentStrategyParameter) map[string]any {
	out := map[string]any{}
	for _, param := range params {
		out[param.Name] = defaultValue(param.Default, string(param.Type), param.Options)
	}
	return out
}

func defaultValue(explicit any, typ string, options []plugin_entities.ParameterOption) any {
	if explicit != nil {
		return explicit
	}
	if len(options) > 0 {
		return options[0].Value
	}
	switch typ {
	case "number":
		return 0
	case "boolean":
		return false
	case "array", "files", "checkbox":
		return []any{}
	case "object":
		return map[string]any{}
	default:
		return ""
	}
}
