package requests

import (
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/model_entities"
)

type BaseRequestInvokeModel struct {
	Provider    string                   `json:"provider"`
	ModelType   model_entities.ModelType `json:"model_type" validate:"required,model_type"`
	Model       string                   `json:"model"`
	Credentials map[string]any           `json:"credentials" validate:"omitempty,dive,is_basic_type"`
}

type RequestInvokeLLM struct {
	BaseRequestInvokeModel

	ModelParameters map[string]any                     `json:"model_parameters" validate:"omitempty,dive,is_basic_type"`
	PromptMessages  []model_entities.PromptMessage     `json:"prompt_messages" validate:"omitempty,dive"`
	Tools           []model_entities.PromptMessageTool `json:"tools" validate:"omitempty,dive"`
	Stop            []string                           `json:"stop" validate:"omitempty"`
	Stream          bool                               `json:"stream"`
}

type RequestInvokeTextEmbedding struct {
	BaseRequestInvokeModel

	Texts []string `json:"texts" validate:"required,dive"`
}

type RequestInvokeRerank struct {
	BaseRequestInvokeModel

	Query          string   `json:"query" validate:"required"`
	Docs           []string `json:"docs" validate:"required,dive"`
	ScoreThreshold float64  `json:"score_threshold"`
	TopN           int      `json:"top_n"`
}

type RequestInvokeTTS struct {
	BaseRequestInvokeModel

	ContentText string `json:"content_text" validate:"required"`
	Voice       string `json:"voice" validate:"required"`
}

type RequestInvokeSpeech2Text struct {
	BaseRequestInvokeModel

	File string `json:"file" validate:"required"` // base64 encoded voice file
}
