package requests

import (
	"encoding/json"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/model_entities"
)

type Credentials struct {
	Credentials    map[string]any `json:"credentials" validate:"omitempty"`
	CredentialType string         `json:"credential_type,omitempty" validate:"omitempty"`
}

type BaseRequestInvokeModel struct {
	Provider string `json:"provider" validate:"required"`
	Model    string `json:"model" validate:"required"`
}

type InvokeLLMSchema struct {
	ModelParameters map[string]any                     `json:"model_parameters"  validate:"omitempty"`
	PromptMessages  []model_entities.PromptMessage     `json:"prompt_messages"  validate:"omitempty"`
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
	Texts     []string `json:"texts" validate:"required,dive"`
	InputType string   `json:"input_type" validate:"required"`
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
	TenantID    string `json:"tenant_id" validate:"required"`
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
	Credentials

	Provider string `json:"provider" validate:"required"`
}

type RequestValidateModelCredentials struct {
	BaseRequestInvokeModel
	Credentials

	ModelType model_entities.ModelType `json:"model_type"  validate:"required,model_type"`
}

type RequestGetTTSModelVoices struct {
	BaseRequestInvokeModel
	Credentials

	ModelType model_entities.ModelType `json:"model_type"  validate:"required,model_type,eq=tts"`
	Language  string                   `json:"language" validate:"omitempty"`
}

type RequestGetTextEmbeddingNumTokens struct {
	BaseRequestInvokeModel
	Credentials

	ModelType model_entities.ModelType `json:"model_type"  validate:"required,model_type,eq=text-embedding"`
	Texts     []string                 `json:"texts" validate:"required,dive"`
}

type RequestGetLLMNumTokens struct {
	BaseRequestInvokeModel
	Credentials

	ModelType      model_entities.ModelType           `json:"model_type"  validate:"required,model_type,eq=llm"`
	PromptMessages []model_entities.PromptMessage     `json:"prompt_messages"  validate:"omitempty,dive"`
	Tools          []model_entities.PromptMessageTool `json:"tools" validate:"omitempty,dive"`
}

func (r RequestGetLLMNumTokens) MarshalJSON() ([]byte, error) {
	type alias RequestGetLLMNumTokens
	p := alias(r)
	if p.PromptMessages == nil {
		p.PromptMessages = []model_entities.PromptMessage{}
	}
	if p.Tools == nil {
		p.Tools = []model_entities.PromptMessageTool{}
	}
	return json.Marshal(p)
}

type RequestGetAIModelSchema struct {
	BaseRequestInvokeModel
	Credentials

	ModelType model_entities.ModelType `json:"model_type"  validate:"required,model_type"`
}
