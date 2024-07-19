package requests

import (
	"encoding/json"

	"github.com/langgenius/dify-plugin-daemon/internal/types/validators"
)

type RequestInvokeTool struct {
	Provider       string         `json:"provider" validate:"required"`
	Tool           string         `json:"tool" validate:"required"`
	ToolParameters map[string]any `json:"tool_parameters" validate:"omitempty,dive,is_basic_type"`
	Credentials    map[string]any `json:"credentials" validate:"omitempty,dive,is_basic_type"`
}

func (r *RequestInvokeTool) UnmarshalJSON(data []byte) error {
	type Alias RequestInvokeTool
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
