package dify_invocation

import (
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/requests"
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
