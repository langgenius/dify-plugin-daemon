package dify_invocation

import (
	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/app_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/internal/types/validators"
)

type BaseInvokeDifyRequest struct {
	TenantId string     `json:"tenant_id"`
	UserId   string     `json:"user_id"`
	Type     InvokeType `json:"type"`
}

type InvokeType string

const (
	INVOKE_TYPE_LLM            InvokeType = "llm"
	INVOKE_TYPE_TEXT_EMBEDDING InvokeType = "text_embedding"
	INVOKE_TYPE_RERANK         InvokeType = "rerank"
	INVOKE_TYPE_TTS            InvokeType = "tts"
	INVOKE_TYPE_SPEECH2TEXT    InvokeType = "speech2text"
	INVOKE_TYPE_MODERATION     InvokeType = "moderation"
	INVOKE_TYPE_TOOL           InvokeType = "tool"
	INVOKE_TYPE_NODE           InvokeType = "node"
	INVOKE_TYPE_APP            InvokeType = "app"
	INVOKE_TYPE_STORAGE        InvokeType = "storage"
)

type InvokeLLMRequest struct {
	BaseInvokeDifyRequest
	requests.BaseRequestInvokeModel
	requests.InvokeLLMSchema
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
	BaseInvokeDifyRequest
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
	AppId          string                  `json:"app_id" validate:"required"`
	Inputs         map[string]any          `json:"inputs" validate:"omitempty"`
	Query          string                  `json:"query" validate:"omitempty"`
	ResponseMode   string                  `json:"response_mode"`
	ConversationId string                  `json:"conversation_id"`
	User           string                  `json:"user" validate:"omitempty"`
	Files          []*app_entities.FileVar `json:"files" validate:"omitempty,dive"`
}

type StorageOpt string

const (
	STORAGE_OPT_GET StorageOpt = "get"
	STORAGE_OPT_SET StorageOpt = "set"
	STORAGE_OPT_DEL StorageOpt = "del"
)

func isStorageOpt(fl validator.FieldLevel) bool {
	opt := StorageOpt(fl.Field().String())
	return opt == STORAGE_OPT_GET || opt == STORAGE_OPT_SET || opt == STORAGE_OPT_DEL
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

type InvokeToolRequest struct {
	BaseInvokeDifyRequest
	ToolType requests.ToolType `json:"tool_type" validate:"required,tool_type"`
	requests.InvokeToolSchema
}

type InvokeNodeResponse struct {
	ProcessData      map[string]any `json:"process_data"`
	Output           map[string]any `json:"output"`
	Input            map[string]any `json:"input"`
	EdgeSourceHandle []string       `json:"edge_source_handle"`
}
