package backwards_invocation

type PluginAccessType string

const (
	PLUGIN_ACCESS_TYPE_TOOL  PluginAccessType = "tool"
	PLUGIN_ACCESS_TYPE_MODEL PluginAccessType = "model"
)

type PluginAccessAction string

const (
	PLUGIN_ACCESS_ACTION_INVOKE_TOOL                   PluginAccessAction = "invoke_tool"
	PLUGIN_ACCESS_ACTION_VALIDATE_TOOL_CREDENTIALS     PluginAccessAction = "validate_tool_credentials"
	PLUGIN_ACCESS_ACTION_INVOKE_LLM                    PluginAccessAction = "invoke_llm"
	PLUGIN_ACCESS_ACTION_INVOKE_TEXT_EMBEDDING         PluginAccessAction = "invoke_text_embedding"
	PLUGIN_ACCESS_ACTION_INVOKE_RERANK                 PluginAccessAction = "invoke_rerank"
	PLUGIN_ACCESS_ACTION_INVOKE_TTS                    PluginAccessAction = "invoke_tts"
	PLUGIN_ACCESS_ACTION_INVOKE_SPEECH2TEXT            PluginAccessAction = "invoke_speech2text"
	PLUGIN_ACCESS_ACTION_INVOKE_MODERATION             PluginAccessAction = "invoke_moderation"
	PLUGIN_ACCESS_ACTION_VALIDATE_PROVIDER_CREDENTIALS PluginAccessAction = "validate_provider_credentials"
	PLUGIN_ACCESS_ACTION_VALIDATE_MODEL_CREDENTIALS    PluginAccessAction = "validate_model_credentials"
)
