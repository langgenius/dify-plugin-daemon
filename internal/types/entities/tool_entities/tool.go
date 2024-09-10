package tool_entities

import (
	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/internal/types/validators"
)

type ToolResponseChunkType string

const (
	ToolResponseChunkTypeText      ToolResponseChunkType = "text"
	ToolResponseChunkTypeFile      ToolResponseChunkType = "file"
	ToolResponseChunkTypeBlob      ToolResponseChunkType = "blob"
	ToolResponseChunkTypeJson      ToolResponseChunkType = "json"
	ToolResponseChunkTypeLink      ToolResponseChunkType = "link"
	ToolResponseChunkTypeImage     ToolResponseChunkType = "image"
	ToolResponseChunkTypeImageLink ToolResponseChunkType = "image_link"
	ToolResponseChunkTypeVariable  ToolResponseChunkType = "variable"
)

func IsValidToolResponseChunkType(fl validator.FieldLevel) bool {
	t := fl.Field().String()
	switch ToolResponseChunkType(t) {
	case ToolResponseChunkTypeText,
		ToolResponseChunkTypeFile,
		ToolResponseChunkTypeBlob,
		ToolResponseChunkTypeJson,
		ToolResponseChunkTypeLink,
		ToolResponseChunkTypeImage,
		ToolResponseChunkTypeImageLink,
		ToolResponseChunkTypeVariable:
		return true
	default:
		return false
	}
}

func init() {
	validators.GlobalEntitiesValidator.RegisterValidation(
		"is_valid_tool_response_chunk_type",
		IsValidToolResponseChunkType,
	)
}

type ToolResponseChunk struct {
	Type    ToolResponseChunkType `json:"type" validate:"required,is_valid_tool_response_chunk_type"`
	Message map[string]any        `json:"message"`
}
