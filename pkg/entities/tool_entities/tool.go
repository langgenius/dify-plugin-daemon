package tool_entities

import (
	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
)

type ToolResponseChunkType string

const (
	ToolResponseChunkTypeText               ToolResponseChunkType = "text"
	ToolResponseChunkTypeFile               ToolResponseChunkType = "file"
	ToolResponseChunkTypeBlob               ToolResponseChunkType = "blob"
	ToolResponseChunkTypeBlobChunk          ToolResponseChunkType = "blob_chunk"
	ToolResponseChunkTypeJson               ToolResponseChunkType = "json"
	ToolResponseChunkTypeLink               ToolResponseChunkType = "link"
	ToolResponseChunkTypeImage              ToolResponseChunkType = "image"
	ToolResponseChunkTypeImageLink          ToolResponseChunkType = "image_link"
	ToolResponseChunkTypeVariable           ToolResponseChunkType = "variable"
	ToolResponseChunkTypeLog                ToolResponseChunkType = "log"
	ToolResponseChunkTypeRetrieverResources ToolResponseChunkType = "retriever_resources"
)

func IsValidToolResponseChunkType(fl validator.FieldLevel) bool {
	t := fl.Field().String()
	switch ToolResponseChunkType(t) {
	case ToolResponseChunkTypeText,
		ToolResponseChunkTypeFile,
		ToolResponseChunkTypeBlob,
		ToolResponseChunkTypeBlobChunk,
		ToolResponseChunkTypeJson,
		ToolResponseChunkTypeLink,
		ToolResponseChunkTypeImage,
		ToolResponseChunkTypeImageLink,
		ToolResponseChunkTypeVariable,
		ToolResponseChunkTypeLog,
		ToolResponseChunkTypeRetrieverResources:
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
	Meta    map[string]any        `json:"meta"`
}

type GetToolRuntimeParametersResponse struct {
	Parameters []plugin_entities.ToolParameter `json:"parameters"`
}
