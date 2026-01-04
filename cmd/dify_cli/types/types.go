package types

type EnvConfig struct {
	FilesURL          string `json:"files_url" validate:"required"`
	InnerAPIURL       string `json:"inner_api_url" validate:"required"`
	InnerAPISessionID string `json:"inner_api_session_id" validate:"required"`
}

type I18nObject struct {
	EnUS   string `json:"en_US" yaml:"en_US"`
	JaJp   string `json:"ja_JP,omitempty" yaml:"ja_JP,omitempty"`
	ZhHans string `json:"zh_Hans,omitempty" yaml:"zh_Hans,omitempty"`
	PtBr   string `json:"pt_BR,omitempty" yaml:"pt_BR,omitempty"`
}

type ToolType string

const (
	ToolTypeBuiltin  ToolType = "builtin"
	ToolTypeWorkflow ToolType = "workflow"
	ToolTypeAPI      ToolType = "api"
	ToolTypeMCP      ToolType = "mcp"
)

type ToolParameterType string

const (
	ToolParameterTypeString  ToolParameterType = "string"
	ToolParameterTypeNumber  ToolParameterType = "number"
	ToolParameterTypeBoolean ToolParameterType = "boolean"
	ToolParameterTypeSelect  ToolParameterType = "select"
	ToolParameterTypeFile    ToolParameterType = "file"
	ToolParameterTypeFiles   ToolParameterType = "files"
)

type InvokeType string

const (
	INVOKE_TYPE_LLM                      InvokeType = "llm"
	INVOKE_TYPE_LLM_STRUCTURED_OUTPUT    InvokeType = "llm_structured_output"
	INVOKE_TYPE_TEXT_EMBEDDING           InvokeType = "text_embedding"
	INVOKE_TYPE_MULTIMODAL_EMBEDDING     InvokeType = "multimodal_embedding"
	INVOKE_TYPE_RERANK                   InvokeType = "rerank"
	INVOKE_TYPE_MULTIMODAL_RERANK        InvokeType = "multimodal_rerank"
	INVOKE_TYPE_TTS                      InvokeType = "tts"
	INVOKE_TYPE_SPEECH2TEXT              InvokeType = "speech2text"
	INVOKE_TYPE_MODERATION               InvokeType = "moderation"
	INVOKE_TYPE_TOOL                     InvokeType = "tool"
	INVOKE_TYPE_NODE_PARAMETER_EXTRACTOR InvokeType = "node_parameter_extractor"
	INVOKE_TYPE_NODE_QUESTION_CLASSIFIER InvokeType = "node_question_classifier"
	INVOKE_TYPE_APP                      InvokeType = "app"
	INVOKE_TYPE_STORAGE                  InvokeType = "storage"
	INVOKE_TYPE_ENCRYPT                  InvokeType = "encrypt"
	INVOKE_TYPE_SYSTEM_SUMMARY           InvokeType = "system_summary"
	INVOKE_TYPE_UPLOAD_FILE              InvokeType = "upload_file"
	INVOKE_TYPE_FETCH_APP                InvokeType = "fetch_app"
)

type ParameterOption struct {
	Value string     `json:"value" yaml:"value"`
	Label I18nObject `json:"label" yaml:"label"`
}

type ToolParameter struct {
	Name             string            `json:"name" yaml:"name"`
	Label            I18nObject        `json:"label" yaml:"label"`
	HumanDescription I18nObject        `json:"human_description" yaml:"human_description"`
	LLMDescription   string            `json:"llm_description" yaml:"llm_description"`
	Type             ToolParameterType `json:"type" yaml:"type"`
	Required         bool              `json:"required" yaml:"required"`
	Default          any               `json:"default" yaml:"default"`
	Options          []ParameterOption `json:"options" yaml:"options"`
}

type ToolDescription struct {
	Human I18nObject `json:"human"`
	LLM   string     `json:"llm"`
}

type ToolOutputSchema map[string]any

type DifyToolIdentity struct {
	Author   string     `json:"author"`
	Name     string     `json:"name"`
	Label    I18nObject `json:"label"`
	Provider string     `json:"provider"`
}

type DifyToolDeclaration struct {
	ProviderType   ToolType         `json:"provider_type" yaml:"provider_type" validate:"required"`
	Identity       DifyToolIdentity `json:"identity" yaml:"identity" validate:"required"`
	Description    ToolDescription  `json:"description" yaml:"description" validate:"required"`
	Parameters     []ToolParameter  `json:"parameters" yaml:"parameters" validate:"omitempty,dive"`
	OutputSchema   ToolOutputSchema `json:"output_schema,omitempty" yaml:"output_schema,omitempty"`
	CredentialType string           `json:"credential_type" yaml:"credential_type" validate:"omitempty"`
	CredentialId   string           `json:"credential_id" yaml:"credential_id" validate:"omitempty"`
}

type DifyConfig struct {
	Env   EnvConfig             `json:"env"`
	Tools []DifyToolDeclaration `json:"tools"`
}

type DifyInnerAPIResponse[T any] struct {
	Data  *T     `json:"data,omitempty"`
	Error string `json:"error"`
}

type DifyToolResponseChunkType string

const (
	ToolResponseChunkTypeBinaryLink         DifyToolResponseChunkType = "binary_link"
	ToolResponseChunkTypeText               DifyToolResponseChunkType = "text"
	ToolResponseChunkTypeFile               DifyToolResponseChunkType = "file"
	ToolResponseChunkTypeBlob               DifyToolResponseChunkType = "blob"
	ToolResponseChunkTypeBlobChunk          DifyToolResponseChunkType = "blob_chunk"
	ToolResponseChunkTypeJson               DifyToolResponseChunkType = "json"
	ToolResponseChunkTypeLink               DifyToolResponseChunkType = "link"
	ToolResponseChunkTypeImage              DifyToolResponseChunkType = "image"
	ToolResponseChunkTypeImageLink          DifyToolResponseChunkType = "image_link"
	ToolResponseChunkTypeVariable           DifyToolResponseChunkType = "variable"
	ToolResponseChunkTypeLog                DifyToolResponseChunkType = "log"
	ToolResponseChunkTypeRetrieverResources DifyToolResponseChunkType = "retriever_resources"
)

type DifyToolResponseChunk struct {
	Type    DifyToolResponseChunkType `json:"type" validate:"required"`
	Message map[string]any            `json:"message" validate:"omitempty"`
	Meta    map[string]any            `json:"meta" validate:"omitempty"`
}

type InvokeToolRequest struct {
	Type           InvokeType     `json:"type"`
	ToolType       ToolType       `json:"tool_type"`
	Provider       string         `json:"provider"`
	Tool           string         `json:"tool"`
	ToolParameters map[string]any `json:"tool_parameters"`
	CredentialId   string         `json:"credential_id"`
	CredentialType string         `json:"credential_type"`
}
