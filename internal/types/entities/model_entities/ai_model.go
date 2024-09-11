package model_entities

import "github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"

type GetModelSchemasResponse struct {
	AIModels []plugin_entities.ModelDeclaration `json:"ai_models" validate:"required,dive"`
}
