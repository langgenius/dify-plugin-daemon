package requests

import (
	"encoding/json"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/validators"
)

type RequestInvokeLLM struct {
	Provider        string                             `json:"provider"`
	ModelType       model_entities.ModelType           `json:"model_type" validate:"required,model_type"`
	Model           string                             `json:"model"`
	ModelParameters map[string]any                     `json:"model_parameters" validate:"omitempty,dive,is_basic_type"`
	PromptMessages  []model_entities.PromptMessage     `json:"prompt_messages" validate:"omitempty,dive"`
	Tools           []model_entities.PromptMessageTool `json:"tools" validate:"omitempty,dive"`
	Stop            []string                           `json:"stop" validate:"omitempty"`
	Stream          bool                               `json:"stream"`
	Credentials     map[string]any                     `json:"credentials" validate:"omitempty,dive,is_basic_type"`
}

func (r *RequestInvokeLLM) UnmarshalJSON(data []byte) error {
	type Alias RequestInvokeLLM
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if err := validators.GlobalEntitiesValidator.Struct(r); err != nil {
		return err
	}

	return nil
}
