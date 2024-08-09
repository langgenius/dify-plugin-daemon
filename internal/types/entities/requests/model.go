package requests

import (
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/model_entities"
)

type Credentials struct {
	Credentials map[string]any `json:"credentials" validate:"omitempty,dive,is_basic_type"`
}

type BaseRequestInvokeModel struct {
	Provider string `json:"provider" validate:"required"`
	Model    string `json:"model" validate:"required"`
}

type InvokeLLMSchema struct {
	Mode            string                             `json:"mode" validate:"required"`
	ModelParameters map[string]any                     `json:"model_parameters"  validate:"omitempty,dive,is_basic_type"`
	PromptMessages  []model_entities.PromptMessage     `json:"prompt_messages"  validate:"omitempty,dive"`
	Tools           []model_entities.PromptMessageTool `json:"tools" validate:"omitempty,dive"`
	Stop            []string                           `json:"stop" validate:"omitempty"`
	Stream          bool                               `json:"stream"`
}

type RequestInvokeLLM struct {
	BaseRequestInvokeModel
	Credentials
	InvokeLLMSchema

	ModelType model_entities.ModelType `json:"model_type"  validate:"required,model_type,eq=llm"`
}

type InvokeTextEmbeddingSchema struct {
	Texts []string `json:"texts" validate:"required,dive"`
}

type RequestInvokeTextEmbedding struct {
	BaseRequestInvokeModel
	Credentials
	InvokeTextEmbeddingSchema

	ModelType model_entities.ModelType `json:"model_type"  validate:"required,model_type,eq=text-embedding"`
}

type InvokeRerankSchema struct {
	Query          string   `json:"query" validate:"required"`
	Docs           []string `json:"docs" validate:"required,dive"`
	ScoreThreshold float64  `json:"score_threshold" `
	TopN           int      `json:"top_n" `
}

type RequestInvokeRerank struct {
	BaseRequestInvokeModel
	Credentials
	InvokeRerankSchema

	ModelType model_entities.ModelType `json:"model_type"  validate:"required,model_type,eq=rerank"`
}

type InvokeTTSSchema struct {
	ContentText string `json:"content_text"  validate:"required"`
	Voice       string `json:"voice" validate:"required"`
}

type RequestInvokeTTS struct {
	BaseRequestInvokeModel
	Credentials
	InvokeTTSSchema

	ModelType model_entities.ModelType `json:"model_type"  validate:"required,model_type,eq=tts"`
}

type InvokeSpeech2TextSchema struct {
	File string `json:"file" validate:"required"` // hexing encoded voice file
}

type RequestInvokeSpeech2Text struct {
	BaseRequestInvokeModel
	Credentials
	InvokeSpeech2TextSchema

	ModelType model_entities.ModelType `json:"model_type"  validate:"required,model_type,eq=speech2text"`
}

type InvokeModerationSchema struct {
	Text string `json:"text" validate:"required"`
}

type RequestInvokeModeration struct {
	BaseRequestInvokeModel
	Credentials
	InvokeModerationSchema

	ModelType model_entities.ModelType `json:"model_type"  validate:"required,model_type,eq=moderation"`
}

type RequestValidateProviderCredentials struct {
	Provider    string         `json:"provider" validate:"required"`
	Credentials map[string]any `json:"credentials" validate:"omitempty,dive,is_basic_type"`
}

type RequestValidateModelCredentials struct {
	Provider    string                   `json:"provider" validate:"required"`
	ModelType   model_entities.ModelType `json:"model_type"  validate:"required,model_type"`
	Model       string                   `json:"model" validate:"required"`
	Credentials map[string]any           `json:"credentials" validate:"omitempty,dive,is_basic_type"`
}
