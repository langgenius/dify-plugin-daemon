package access_types

type PluginAccessType string

const (
	PLUGIN_ACCESS_TYPE_TOOL           PluginAccessType = "tool"
	PLUGIN_ACCESS_TYPE_MODEL          PluginAccessType = "model"
	PLUGIN_ACCESS_TYPE_ENDPOINT       PluginAccessType = "endpoint"
	PLUGIN_ACCESS_TYPE_AGENT_STRATEGY PluginAccessType = "agent_strategy"
)

func (p PluginAccessType) IsValid() bool {
	return p == PLUGIN_ACCESS_TYPE_TOOL ||
		p == PLUGIN_ACCESS_TYPE_MODEL ||
		p == PLUGIN_ACCESS_TYPE_ENDPOINT ||
		p == PLUGIN_ACCESS_TYPE_AGENT_STRATEGY
}

type PluginAccessAction string

const (
	PLUGIN_ACCESS_ACTION_INVOKE_TOOL                   PluginAccessAction = "invoke_tool"
	PLUGIN_ACCESS_ACTION_VALIDATE_TOOL_CREDENTIALS     PluginAccessAction = "validate_tool_credentials"
	PLUGIN_ACCESS_ACTION_GET_TOOL_RUNTIME_PARAMETERS   PluginAccessAction = "get_tool_runtime_parameters"
	PLUGIN_ACCESS_ACTION_INVOKE_LLM                    PluginAccessAction = "invoke_llm"
	PLUGIN_ACCESS_ACTION_INVOKE_TEXT_EMBEDDING         PluginAccessAction = "invoke_text_embedding"
	PLUGIN_ACCESS_ACTION_INVOKE_RERANK                 PluginAccessAction = "invoke_rerank"
	PLUGIN_ACCESS_ACTION_INVOKE_TTS                    PluginAccessAction = "invoke_tts"
	PLUGIN_ACCESS_ACTION_INVOKE_SPEECH2TEXT            PluginAccessAction = "invoke_speech2text"
	PLUGIN_ACCESS_ACTION_INVOKE_MODERATION             PluginAccessAction = "invoke_moderation"
	PLUGIN_ACCESS_ACTION_VALIDATE_PROVIDER_CREDENTIALS PluginAccessAction = "validate_provider_credentials"
	PLUGIN_ACCESS_ACTION_VALIDATE_MODEL_CREDENTIALS    PluginAccessAction = "validate_model_credentials"
	PLUGIN_ACCESS_ACTION_INVOKE_ENDPOINT               PluginAccessAction = "invoke_endpoint"
	PLUGIN_ACCESS_ACTION_GET_TTS_MODEL_VOICES          PluginAccessAction = "get_tts_model_voices"
	PLUGIN_ACCESS_ACTION_GET_TEXT_EMBEDDING_NUM_TOKENS PluginAccessAction = "get_text_embedding_num_tokens"
	PLUGIN_ACCESS_ACTION_GET_AI_MODEL_SCHEMAS          PluginAccessAction = "get_ai_model_schemas"
	PLUGIN_ACCESS_ACTION_GET_LLM_NUM_TOKENS            PluginAccessAction = "get_llm_num_tokens"
	PLUGIN_ACCESS_ACTION_INVOKE_AGENT_STRATEGY         PluginAccessAction = "invoke_agent_strategy"
)

func (p PluginAccessAction) IsValid() bool {
	return p == PLUGIN_ACCESS_ACTION_INVOKE_TOOL ||
		p == PLUGIN_ACCESS_ACTION_VALIDATE_TOOL_CREDENTIALS ||
		p == PLUGIN_ACCESS_ACTION_GET_TOOL_RUNTIME_PARAMETERS ||
		p == PLUGIN_ACCESS_ACTION_INVOKE_LLM ||
		p == PLUGIN_ACCESS_ACTION_INVOKE_TEXT_EMBEDDING ||
		p == PLUGIN_ACCESS_ACTION_INVOKE_RERANK ||
		p == PLUGIN_ACCESS_ACTION_INVOKE_TTS ||
		p == PLUGIN_ACCESS_ACTION_INVOKE_SPEECH2TEXT ||
		p == PLUGIN_ACCESS_ACTION_INVOKE_MODERATION ||
		p == PLUGIN_ACCESS_ACTION_VALIDATE_PROVIDER_CREDENTIALS ||
		p == PLUGIN_ACCESS_ACTION_VALIDATE_MODEL_CREDENTIALS ||
		p == PLUGIN_ACCESS_ACTION_INVOKE_ENDPOINT ||
		p == PLUGIN_ACCESS_ACTION_GET_TTS_MODEL_VOICES ||
		p == PLUGIN_ACCESS_ACTION_GET_TEXT_EMBEDDING_NUM_TOKENS ||
		p == PLUGIN_ACCESS_ACTION_GET_AI_MODEL_SCHEMAS ||
		p == PLUGIN_ACCESS_ACTION_GET_LLM_NUM_TOKENS ||
		p == PLUGIN_ACCESS_ACTION_INVOKE_AGENT_STRATEGY
}
