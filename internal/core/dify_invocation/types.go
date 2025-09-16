package dify_invocation

import (
	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
)

type BaseInvokeDifyRequest struct {
	TenantId string     `json:"tenant_id"`
	UserId   string     `json:"user_id"`
	Type     InvokeType `json:"type"`
}

type InvokeType string

const (
	INVOKE_TYPE_LLM                      InvokeType = "llm"
	INVOKE_TYPE_LLM_STRUCTURED_OUTPUT    InvokeType = "llm_structured_output"
	INVOKE_TYPE_TEXT_EMBEDDING           InvokeType = "text_embedding"
	INVOKE_TYPE_RERANK                   InvokeType = "rerank"
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

type InvokeLLMSchema struct {
	Mode             string                             `json:"mode" validate:"required"`
	CompletionParams map[string]any                     `json:"completion_params"  validate:"omitempty"`
	PromptMessages   []model_entities.PromptMessage     `json:"prompt_messages"  validate:"omitempty"`
	Tools            []model_entities.PromptMessageTool `json:"tools" validate:"omitempty,dive"`
	Stop             []string                           `json:"stop" validate:"omitempty"`
	Stream           bool                               `json:"stream"`
}

type InvokeLLMRequest struct {
	BaseInvokeDifyRequest
	requests.BaseRequestInvokeModel
	// requests.InvokeLLMSchema,
	// TODO: as completion_params in requests.InvokeLLMSchema is "model_parameters", we declare another one here
	InvokeLLMSchema
}

type InvokeLLMWithStructuredOutputRequest struct {
	BaseInvokeDifyRequest
	requests.BaseRequestInvokeModel
	// requests.InvokeLLMSchema
	// TODO: as completion_params in requests.InvokeLLMSchema is "model_parameters", we declare another one here
	InvokeLLMSchema
	StructuredOutputSchema map[string]any `json:"structured_output_schema" validate:"required"`
}

type InvokeTextEmbeddingRequest struct {
	BaseInvokeDifyRequest
	requests.BaseRequestInvokeModel
	requests.InvokeTextEmbeddingSchema
}

type InvokeRerankRequest struct {
	BaseInvokeDifyRequest
	requests.BaseRequestInvokeModel
	requests.InvokeRerankSchema
}

type InvokeTTSRequest struct {
	// BaseInvokeDifyRequest
	// # TODO: BaseInvokeDifyRequest has a duplicate field with InvokeTTSSchema,
	// # we should consider to refactor it in the future
	UserId string     `json:"user_id"`
	Type   InvokeType `json:"type"`
	requests.BaseRequestInvokeModel
	requests.InvokeTTSSchema
}

type InvokeSpeech2TextRequest struct {
	BaseInvokeDifyRequest
	requests.BaseRequestInvokeModel
	requests.InvokeSpeech2TextSchema
}

type InvokeModerationRequest struct {
	BaseInvokeDifyRequest
	requests.BaseRequestInvokeModel
	requests.InvokeModerationSchema
}

type InvokeAppSchema struct {
	AppId          string         `json:"app_id" validate:"required"`
	Inputs         map[string]any `json:"inputs" validate:"omitempty"`
	Query          string         `json:"query" validate:"omitempty"`
	ResponseMode   string         `json:"response_mode"`
	ConversationId string         `json:"conversation_id" validate:"omitempty"`
	User           string         `json:"user" validate:"omitempty"`
}

type StorageOpt string

const (
	STORAGE_OPT_GET   StorageOpt = "get"
	STORAGE_OPT_SET   StorageOpt = "set"
	STORAGE_OPT_DEL   StorageOpt = "del"
	STORAGE_OPT_EXIST StorageOpt = "exist"
)

func isStorageOpt(fl validator.FieldLevel) bool {
	opt := StorageOpt(fl.Field().String())
	return opt == STORAGE_OPT_GET || opt == STORAGE_OPT_SET || opt == STORAGE_OPT_DEL || opt == STORAGE_OPT_EXIST
}

func init() {
	validators.GlobalEntitiesValidator.RegisterValidation("storage_opt", isStorageOpt)
}

type InvokeStorageRequest struct {
	Opt   StorageOpt `json:"opt" validate:"required,storage_opt"`
	Key   string     `json:"key" validate:"required"`
	Value string     `json:"value"` // encoded in hex, optional
}

type InvokeAppRequest struct {
	BaseInvokeDifyRequest

	InvokeAppSchema
}

type ModelConfig struct {
	Provider         string         `json:"provider" validate:"required"`
	Name             string         `json:"name" validate:"required"`
	Mode             string         `json:"mode" validate:"required"`
	CompletionParams map[string]any `json:"completion_params" validate:"omitempty"`
}

type InvokeParameterExtractorRequest struct {
	BaseInvokeDifyRequest

	Parameters []struct {
		Name        string   `json:"name" validate:"required"`
		Type        string   `json:"type" validate:"required,oneof=string number bool select array[string] array[number] array[object]"`
		Options     []string `json:"options" validate:"omitempty"`
		Description string   `json:"description" validate:"omitempty"`
		Required    bool     `json:"required" validate:"omitempty"`
	} `json:"parameters" validate:"required,dive"`

	Model       ModelConfig `json:"model" validate:"required"`
	Instruction string      `json:"instruction" validate:"omitempty"`
	Query       string      `json:"query" validate:"required"`
}

type InvokeQuestionClassifierRequest struct {
	BaseInvokeDifyRequest

	Classes []struct {
		ID   string `json:"id" validate:"required"`
		Name string `json:"name" validate:"required"`
	} `json:"classes" validate:"required,dive"`

	Model       ModelConfig `json:"model" validate:"required"`
	Instruction string      `json:"instruction" validate:"omitempty"`
	Query       string      `json:"query" validate:"required"`
}

type EncryptOpt string

const (
	ENCRYPT_OPT_ENCRYPT EncryptOpt = "encrypt"
	ENCRYPT_OPT_DECRYPT EncryptOpt = "decrypt"
	ENCRYPT_OPT_CLEAR   EncryptOpt = "clear"
)

func isEncryptOpt(fl validator.FieldLevel) bool {
	opt := EncryptOpt(fl.Field().String())
	return opt == ENCRYPT_OPT_ENCRYPT || opt == ENCRYPT_OPT_DECRYPT || opt == ENCRYPT_OPT_CLEAR
}

type EncryptNamespace string

const (
	ENCRYPT_NAMESPACE_ENDPOINT EncryptNamespace = "endpoint"
)

func isEncryptNamespace(fl validator.FieldLevel) bool {
	opt := EncryptNamespace(fl.Field().String())
	return opt == ENCRYPT_NAMESPACE_ENDPOINT
}

func init() {
	validators.GlobalEntitiesValidator.RegisterValidation("encrypt_opt", isEncryptOpt)
	validators.GlobalEntitiesValidator.RegisterValidation("encrypt_namespace", isEncryptNamespace)
}

type InvokeEncryptSchema struct {
	Opt       EncryptOpt                       `json:"opt" validate:"required,encrypt_opt"`
	Namespace EncryptNamespace                 `json:"namespace" validate:"required,encrypt_namespace"`
	Identity  string                           `json:"identity" validate:"required"`
	Data      map[string]any                   `json:"data" validate:"omitempty"`
	Config    []plugin_entities.ProviderConfig `json:"config" validate:"omitempty,dive"`
}

type InvokeEncryptRequest struct {
	BaseInvokeDifyRequest

	InvokeEncryptSchema
}

func (r *InvokeEncryptRequest) EncryptRequired(settings map[string]any) bool {
	if r.Config == nil {
		return false
	}

	// filter out which key needs encrypt
	for _, config := range r.Config {
		if config.Type == plugin_entities.CONFIG_TYPE_SECRET_INPUT {
			return true
		}
	}

	return false
}

type InvokeToolRequest struct {
	BaseInvokeDifyRequest
	ToolType       requests.ToolType `json:"tool_type" validate:"required,tool_type"`
	CredentialId   string            `json:"credential_id" validate:"omitempty"`
	CredentialType string            `json:"credential_type,omitempty" validate:"omitempty"`
	requests.InvokeToolSchema
}

type InvokeNodeResponse struct {
	ProcessData map[string]any `json:"process_data" validate:"required"`
	Outputs     map[string]any `json:"outputs" validate:"required"`
	Inputs      map[string]any `json:"inputs" validate:"required"`
}

type InvokeEncryptionResponse struct {
	Error string         `json:"error"`
	Data  map[string]any `json:"data"`
}

type InvokeSummarySchema struct {
	Text        string `json:"text" validate:"required"`
	Instruction string `json:"instruction" validate:"omitempty"`
}

type InvokeSummaryRequest struct {
	BaseInvokeDifyRequest
	InvokeSummarySchema
}

type InvokeSummaryResponse struct {
	Summary string `json:"summary"`
}

type UploadFileRequest struct {
	BaseInvokeDifyRequest
	Filename string `json:"filename" validate:"required"`
	MimeType string `json:"mimetype" validate:"required"`
}

type UploadFileResponse struct {
	URL string `json:"url"`
}

type FetchAppRequest struct {
	BaseInvokeDifyRequest
	AppId string `json:"app_id" validate:"required"`
}
