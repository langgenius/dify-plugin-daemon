package plugin_daemon

type PluginAccessType string

const (
	PLUGIN_ACCESS_TYPE_TOOL  PluginAccessType = "tool"
	PLUGIN_ACCESS_TYPE_MODEL PluginAccessType = "model"
)

type PluginAccessAction string

const (
	PLUGIN_ACCESS_ACTION_INVOKE_TOOL           PluginAccessAction = "invoke_tool"
	PLUGIN_ACCESS_ACTION_INVOKE_LLM            PluginAccessAction = "invoke_llm"
	PLUGIN_ACCESS_ACTION_INVOKE_TEXT_EMBEDDING PluginAccessAction = "invoke_text_embedding"
	PLUGIN_ACCESS_ACTION_INVOKE_RERANK         PluginAccessAction = "invoke_rerank"
	PLUGIN_ACCESS_ACTION_INVOKE_TTS            PluginAccessAction = "invoke_tts"
	PLUGIN_ACCESS_ACTION_INVOKE_SPEECH2TEXT    PluginAccessAction = "invoke_speech2text"
	PLUGIN_ACCESS_ACTION_INVOKE_MODERATION     PluginAccessAction = "invoke_moderation"
)

const (
	PLUGIN_IN_STREAM_EVENT = "request"
)

func getBasicPluginAccessMap(session_id string, user_id string, access_type PluginAccessType, action PluginAccessAction) map[string]any {
	return map[string]any{
		"session_id": session_id,
		"event":      PLUGIN_IN_STREAM_EVENT,
		"data": map[string]any{
			"user_id": user_id,
			"type":    access_type,
			"action":  action,
		},
	}
}
