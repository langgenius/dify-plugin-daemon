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
	INVOKE_TYPE_LLM            InvokeType = "LLM"
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
	Data struct {
		requests.BaseRequestInvokeModel
		requests.InvokeLLMSchema
	} `json:"data" validate:"required"`
}

type InvokeTextEmbeddingRequest struct {
	BaseInvokeDifyRequest
	Data struct {
		requests.BaseRequestInvokeModel
		requests.InvokeTextEmbeddingSchema
	} `json:"data" validate:"required"`
}

type InvokeRerankRequest struct {
	BaseInvokeDifyRequest
	Data struct {
		requests.BaseRequestInvokeModel
		requests.InvokeRerankSchema
	} `json:"data" validate:"required"`
}

type InvokeTTSRequest struct {
	BaseInvokeDifyRequest
	Data struct {
		requests.BaseRequestInvokeModel
		requests.InvokeTTSSchema
	} `json:"data" validate:"required"`
}

type InvokeSpeech2TextRequest struct {
	BaseInvokeDifyRequest
	Data struct {
		requests.BaseRequestInvokeModel
		requests.InvokeSpeech2TextSchema
	} `json:"data" validate:"required"`
}

type InvokeModerationRequest struct {
	BaseInvokeDifyRequest
	Data struct {
		requests.BaseRequestInvokeModel
		requests.InvokeModerationSchema
	} `json:"data" validate:"required"`
}

type InvokeToolRequest struct {
	BaseInvokeDifyRequest
	Data struct {
		requests.RequestInvokeTool
	} `json:"data" validate:"required"`
}

type InvokeNodeResponse struct {
	ProcessData      map[string]any `json:"process_data"`
	Output           map[string]any `json:"output"`
	Input            map[string]any `json:"input"`
	EdgeSourceHandle []string       `json:"edge_source_handle"`
}
