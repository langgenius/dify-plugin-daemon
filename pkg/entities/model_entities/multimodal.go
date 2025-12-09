package model_entities

import (
	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
)

type MultimodalContentType string

const (
	MULTIMODAL_CONTENT_TYPE_TEXT  MultimodalContentType = "text"
	MULTIMODAL_CONTENT_TYPE_IMAGE MultimodalContentType = "image"
)

type MultimodalContent struct {
	Content     string                `json:"content" validate:"required"`
	ContentType MultimodalContentType `json:"content_type" validate:"required,multimodal_content_type"`
}

type MultimodalRerankDocument struct {
	Index *int     `json:"index" validate:"required"`
	Text  *string  `json:"text" validate:"required"`
	Score *float64 `json:"score" validate:"required"`
}

type MultimodalRerankResult struct {
	Model string                     `json:"model" validate:"required"`
	Docs  []MultimodalRerankDocument `json:"docs" validate:"required,dive"`
}

type MultimodalEmbeddingResult struct {
	Model      string         `json:"model" validate:"required"`
	Embeddings [][]float64    `json:"embeddings" validate:"required,dive"`
	Usage      EmbeddingUsage `json:"usage" validate:"required"`
}

func isMultimodalContentType(fl validator.FieldLevel) bool {
	value := MultimodalContentType(fl.Field().String())
	switch value {
	case MULTIMODAL_CONTENT_TYPE_TEXT, MULTIMODAL_CONTENT_TYPE_IMAGE:
		return true
	}
	return false
}

func init() {
	validators.GlobalEntitiesValidator.RegisterValidation("multimodal_content_type", isMultimodalContentType)
}
