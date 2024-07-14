package plugin_entities

import "testing"

func TestGenericType_Validate(t *testing.T) {
	type F struct {
		G map[string]any `json:"g" validate:"omitempty,dive,is_basic_type"`
	}

	f := F{
		G: map[string]any{
			"key": "value",
		},
	}

	if err := global_tool_provider_validator.Struct(f); err != nil {
		t.Errorf("GenericType_Validate() error = %v", err)
	}
}

func TestWrongGenericType_Validate(t *testing.T) {
	type F struct {
		G map[string]any `json:"g" validate:"omitempty,dive,is_basic_type"`
	}

	f := F{
		G: map[string]any{
			"key": struct{}{},
		},
	}

	if err := global_tool_provider_validator.Struct(f); err == nil {
		t.Error("WrongGenericType_Validate() error = nil; want error")
	}
}
