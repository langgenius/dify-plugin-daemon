package dynamic_select_entities

import "github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"

type DynamicSelectResult struct {
	Options []plugin_entities.ParameterOption `json:"options" validate:"omitempty,dive"`
}
