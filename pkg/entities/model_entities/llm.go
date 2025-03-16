package model_entities

import (
	"encoding/json"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
	"github.com/shopspring/decimal"
)

type ModelType string

const (
	MODEL_TYPE_LLM            ModelType = "llm"
	MODEL_TYPE_TEXT_EMBEDDING ModelType = "text-embedding"
	MODEL_TYPE_RERANKING      ModelType = "rerank"
	MODEL_TYPE_SPEECH2TEXT    ModelType = "speech2text"
	MODEL_TYPE_TTS            ModelType = "tts"
	MODEL_TYPE_MODERATION     ModelType = "moderation"
)

type LLMModel string

const (
	LLM_MODE_CHAT       LLMModel = "chat"
	LLM_MODE_COMPLETION LLMModel = "completion"
)

type PromptMessageRole string

const (
	PROMPT_MESSAGE_ROLE_SYSTEM    = "system"
	PROMPT_MESSAGE_ROLE_USER      = "user"
	PROMPT_MESSAGE_ROLE_ASSISTANT = "assistant"
	PROMPT_MESSAGE_ROLE_TOOL      = "tool"
)

func isPromptMessageRole(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case string(PROMPT_MESSAGE_ROLE_SYSTEM),
		string(PROMPT_MESSAGE_ROLE_USER),
		string(PROMPT_MESSAGE_ROLE_ASSISTANT),
		string(PROMPT_MESSAGE_ROLE_TOOL):
		return true
	}
	return false
}

type PromptMessage struct {
	Role       PromptMessageRole       `json:"role" validate:"required,prompt_message_role"`
	Content    any                     `json:"content" validate:"required,prompt_message_content"`
	Name       string                  `json:"name"`
	ToolCalls  []PromptMessageToolCall `json:"tool_calls" validate:"dive"`
	ToolCallId string                  `json:"tool_call_id"`
}

func isPromptMessageContent(fl validator.FieldLevel) bool {
	// only allow string or []PromptMessageContent
	value := fl.Field().Interface()
	switch valueType := value.(type) {
	case string:
		return true
	case []PromptMessageContent:
		// validate the content
		for _, content := range valueType {
			if err := validators.GlobalEntitiesValidator.Struct(content); err != nil {
				return false
			}
		}
		return true
	}
	return false
}

type PromptMessageContentType string

const (
	PROMPT_MESSAGE_CONTENT_TYPE_TEXT     PromptMessageContentType = "text"
	PROMPT_MESSAGE_CONTENT_TYPE_IMAGE    PromptMessageContentType = "image"
	PROMPT_MESSAGE_CONTENT_TYPE_AUDIO    PromptMessageContentType = "audio"
	PROMPT_MESSAGE_CONTENT_TYPE_VIDEO    PromptMessageContentType = "video"
	PROMPT_MESSAGE_CONTENT_TYPE_DOCUMENT PromptMessageContentType = "document"
)

func isPromptMessageContentType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case string(PROMPT_MESSAGE_CONTENT_TYPE_TEXT),
		string(PROMPT_MESSAGE_CONTENT_TYPE_IMAGE),
		string(PROMPT_MESSAGE_CONTENT_TYPE_AUDIO),
		string(PROMPT_MESSAGE_CONTENT_TYPE_VIDEO),
		string(PROMPT_MESSAGE_CONTENT_TYPE_DOCUMENT):
		return true
	}
	return false
}

type PromptMessageContent struct {
	Type         PromptMessageContentType `json:"type" validate:"required,prompt_message_content_type"`
	Base64Data   string                   `json:"base64_data"` // for multi-modal data
	URL          string                   `json:"url"`         // for multi-modal data
	Data         string                   `json:"data"`        // for text only
	EncodeFormat string                   `json:"encode_format"`
	Format       string                   `json:"format"`
	MimeType     string                   `json:"mime_type"`
	Detail       string                   `json:"detail"`      // for multi-modal data
}

type PromptMessageToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

func init() {
	validators.GlobalEntitiesValidator.RegisterValidation("prompt_message_role", isPromptMessageRole)
	validators.GlobalEntitiesValidator.RegisterValidation("prompt_message_content", isPromptMessageContent)
	validators.GlobalEntitiesValidator.RegisterValidation("prompt_message_content_type", isPromptMessageContentType)
}

func (p *PromptMessage) UnmarshalJSON(data []byte) error {
	// Unmarshal the JSON data into a map
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Check if content is a string or an array which contains type and content
	if _, ok := raw["content"]; ok {
		var content string
		if err := json.Unmarshal(raw["content"], &content); err == nil {
			p.Content = content
		} else {
			var content []PromptMessageContent
			if err := json.Unmarshal(raw["content"], &content); err != nil {
				return err
			}
			p.Content = content
		}
	} else {
		return errors.New("content field is required")
	}

	// Unmarshal the rest of the fields
	if role, ok := raw["role"]; ok {
		if err := json.Unmarshal(role, &p.Role); err != nil {
			return err
		}
	} else {
		return errors.New("role field is required")
	}

	if name, ok := raw["name"]; ok {
		if err := json.Unmarshal(name, &p.Name); err != nil {
			return err
		}
	}

	if toolCalls, ok := raw["tool_calls"]; ok {
		if err := json.Unmarshal(toolCalls, &p.ToolCalls); err != nil {
			return err
		}
	}

	if toolCallId, ok := raw["tool_call_id"]; ok {
		if err := json.Unmarshal(toolCallId, &p.ToolCallId); err != nil {
			return err
		}
	}

	return nil
}

type PromptMessageTool struct {
	Name        string         `json:"name" validate:"required"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type LLMResultChunk struct {
	Model             LLMModel            `json:"model" validate:"required"`
	PromptMessages    []PromptMessage     `json:"prompt_messages" validate:"required,dive"`
	SystemFingerprint string              `json:"system_fingerprint" validate:"omitempty"`
	Delta             LLMResultChunkDelta `json:"delta" validate:"required"`
}

type LLMUsage struct {
	PromptTokens        *int            `json:"prompt_tokens" validate:"required"`
	PromptUnitPrice     decimal.Decimal `json:"prompt_unit_price" validate:"required"`
	PromptPriceUnit     decimal.Decimal `json:"prompt_price_unit" validate:"required"`
	PromptPrice         decimal.Decimal `json:"prompt_price" validate:"required"`
	CompletionTokens    *int            `json:"completion_tokens" validate:"required"`
	CompletionUnitPrice decimal.Decimal `json:"completion_unit_price" validate:"required"`
	CompletionPriceUnit decimal.Decimal `json:"completion_price_unit" validate:"required"`
	CompletionPrice     decimal.Decimal `json:"completion_price" validate:"required"`
	TotalTokens         *int            `json:"total_tokens" validate:"required"`
	TotalPrice          decimal.Decimal `json:"total_price" validate:"required"`
	Currency            *string         `json:"currency" validate:"required"`
	Latency             *float64        `json:"latency" validate:"required"`
}

type LLMResultChunkDelta struct {
	Index        *int          `json:"index" validate:"required"`
	Message      PromptMessage `json:"message" validate:"required"`
	Usage        *LLMUsage     `json:"usage" validate:"omitempty"`
	FinishReason *string       `json:"finish_reason" validate:"omitempty"`
}

type LLMGetNumTokensResponse struct {
	NumTokens int `json:"num_tokens"`
}
