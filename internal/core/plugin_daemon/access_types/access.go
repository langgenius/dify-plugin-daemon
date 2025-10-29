package access_types

type PluginAccessType string

const (
	PLUGIN_ACCESS_TYPE_TOOL              PluginAccessType = "tool"
	PLUGIN_ACCESS_TYPE_MODEL             PluginAccessType = "model"
	PLUGIN_ACCESS_TYPE_ENDPOINT          PluginAccessType = "endpoint"
	PLUGIN_ACCESS_TYPE_AGENT_STRATEGY    PluginAccessType = "agent_strategy"
	PLUGIN_ACCESS_TYPE_OAUTH             PluginAccessType = "oauth"
	PLUGIN_ACCESS_TYPE_DATASOURCE        PluginAccessType = "datasource"
	PLUGIN_ACCESS_TYPE_DYNAMIC_PARAMETER PluginAccessType = "dynamic_parameter"
	PLUGIN_ACCESS_TYPE_TRIGGER           PluginAccessType = "trigger"
)

func (p PluginAccessType) IsValid() bool {
	return p == PLUGIN_ACCESS_TYPE_TOOL ||
		p == PLUGIN_ACCESS_TYPE_MODEL ||
		p == PLUGIN_ACCESS_TYPE_ENDPOINT ||
		p == PLUGIN_ACCESS_TYPE_AGENT_STRATEGY ||
		p == PLUGIN_ACCESS_TYPE_OAUTH ||
		p == PLUGIN_ACCESS_TYPE_DATASOURCE ||
		p == PLUGIN_ACCESS_TYPE_DYNAMIC_PARAMETER ||
		p == PLUGIN_ACCESS_TYPE_TRIGGER
}

type PluginAccessAction string

const (
	PLUGIN_ACCESS_ACTION_INVOKE_TOOL                                        PluginAccessAction = "invoke_tool"
	PLUGIN_ACCESS_ACTION_VALIDATE_TOOL_CREDENTIALS                          PluginAccessAction = "validate_tool_credentials"
	PLUGIN_ACCESS_ACTION_GET_TOOL_RUNTIME_PARAMETERS                        PluginAccessAction = "get_tool_runtime_parameters"
	PLUGIN_ACCESS_ACTION_INVOKE_LLM                                         PluginAccessAction = "invoke_llm"
	PLUGIN_ACCESS_ACTION_INVOKE_TEXT_EMBEDDING                              PluginAccessAction = "invoke_text_embedding"
	PLUGIN_ACCESS_ACTION_INVOKE_RERANK                                      PluginAccessAction = "invoke_rerank"
	PLUGIN_ACCESS_ACTION_INVOKE_TTS                                         PluginAccessAction = "invoke_tts"
	PLUGIN_ACCESS_ACTION_INVOKE_SPEECH2TEXT                                 PluginAccessAction = "invoke_speech2text"
	PLUGIN_ACCESS_ACTION_INVOKE_MODERATION                                  PluginAccessAction = "invoke_moderation"
	PLUGIN_ACCESS_ACTION_VALIDATE_PROVIDER_CREDENTIALS                      PluginAccessAction = "validate_provider_credentials"
	PLUGIN_ACCESS_ACTION_VALIDATE_MODEL_CREDENTIALS                         PluginAccessAction = "validate_model_credentials"
	PLUGIN_ACCESS_ACTION_INVOKE_ENDPOINT                                    PluginAccessAction = "invoke_endpoint"
	PLUGIN_ACCESS_ACTION_GET_TTS_MODEL_VOICES                               PluginAccessAction = "get_tts_model_voices"
	PLUGIN_ACCESS_ACTION_GET_TEXT_EMBEDDING_NUM_TOKENS                      PluginAccessAction = "get_text_embedding_num_tokens"
	PLUGIN_ACCESS_ACTION_GET_AI_MODEL_SCHEMAS                               PluginAccessAction = "get_ai_model_schemas"
	PLUGIN_ACCESS_ACTION_GET_LLM_NUM_TOKENS                                 PluginAccessAction = "get_llm_num_tokens"
	PLUGIN_ACCESS_ACTION_INVOKE_AGENT_STRATEGY                              PluginAccessAction = "invoke_agent_strategy"
	PLUGIN_ACCESS_ACTION_GET_AUTHORIZATION_URL                              PluginAccessAction = "get_authorization_url"
	PLUGIN_ACCESS_ACTION_GET_CREDENTIALS                                    PluginAccessAction = "get_credentials"
	PLUGIN_ACCESS_ACTION_REFRESH_CREDENTIALS                                PluginAccessAction = "refresh_credentials"
	PLUGIN_ACCESS_ACTION_VALIDATE_CREDENTIALS                               PluginAccessAction = "validate_datasource_credentials"
	PLUGIN_ACCESS_ACTION_INVOKE_WEBSITE_DATASOURCE_GET_CRAWL                PluginAccessAction = "invoke_website_datasource_get_crawl"
	PLUGIN_ACCESS_ACTION_INVOKE_ONLINE_DOCUMENT_DATASOURCE_GET_PAGES        PluginAccessAction = "invoke_online_document_datasource_get_pages"
	PLUGIN_ACCESS_ACTION_INVOKE_ONLINE_DOCUMENT_DATASOURCE_GET_PAGE_CONTENT PluginAccessAction = "invoke_online_document_datasource_get_page_content"
	PLUGIN_ACCESS_ACTION_INVOKE_ONLINE_DRIVE_BROWSE_FILES                   PluginAccessAction = "invoke_online_drive_browse_files"
	PLUGIN_ACCESS_ACTION_INVOKE_ONLINE_DRIVE_DOWNLOAD_FILE                  PluginAccessAction = "invoke_online_drive_download_file"
	PLUGIN_ACCESS_ACTION_DYNAMIC_PARAMETER_FETCH_OPTIONS                    PluginAccessAction = "fetch_parameter_options"
	PLUGIN_ACCESS_ACTION_INVOKE_TRIGGER_EVENT                               PluginAccessAction = "invoke_trigger_event"
	PLUGIN_ACCESS_ACTION_DISPATCH_TRIGGER_EVENT                             PluginAccessAction = "dispatch_trigger_event"
	PLUGIN_ACCESS_ACTION_SUBSCRIBE_TRIGGER                                  PluginAccessAction = "subscribe_trigger"
	PLUGIN_ACCESS_ACTION_UNSUBSCRIBE_TRIGGER                                PluginAccessAction = "unsubscribe_trigger"
	PLUGIN_ACCESS_ACTION_REFRESH_TRIGGER                                    PluginAccessAction = "refresh_trigger"
	PLUGIN_ACCESS_ACTION_VALIDATE_TRIGGER_CREDENTIALS                       PluginAccessAction = "validate_trigger_credentials"
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
		p == PLUGIN_ACCESS_ACTION_INVOKE_AGENT_STRATEGY ||
		p == PLUGIN_ACCESS_ACTION_GET_AUTHORIZATION_URL ||
		p == PLUGIN_ACCESS_ACTION_GET_CREDENTIALS ||
		p == PLUGIN_ACCESS_ACTION_REFRESH_CREDENTIALS ||
		p == PLUGIN_ACCESS_ACTION_DYNAMIC_PARAMETER_FETCH_OPTIONS ||
		p == PLUGIN_ACCESS_ACTION_VALIDATE_CREDENTIALS ||
		p == PLUGIN_ACCESS_ACTION_INVOKE_WEBSITE_DATASOURCE_GET_CRAWL ||
		p == PLUGIN_ACCESS_ACTION_INVOKE_ONLINE_DOCUMENT_DATASOURCE_GET_PAGES ||
		p == PLUGIN_ACCESS_ACTION_INVOKE_ONLINE_DOCUMENT_DATASOURCE_GET_PAGE_CONTENT ||
		p == PLUGIN_ACCESS_ACTION_INVOKE_ONLINE_DRIVE_BROWSE_FILES ||
		p == PLUGIN_ACCESS_ACTION_INVOKE_ONLINE_DRIVE_DOWNLOAD_FILE ||
		p == PLUGIN_ACCESS_ACTION_INVOKE_TRIGGER_EVENT ||
		p == PLUGIN_ACCESS_ACTION_DISPATCH_TRIGGER_EVENT ||
		p == PLUGIN_ACCESS_ACTION_SUBSCRIBE_TRIGGER ||
		p == PLUGIN_ACCESS_ACTION_UNSUBSCRIBE_TRIGGER ||
		p == PLUGIN_ACCESS_ACTION_REFRESH_TRIGGER ||
		p == PLUGIN_ACCESS_ACTION_VALIDATE_TRIGGER_CREDENTIALS
}
