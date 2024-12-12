package plugin_entities

import (
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

func TestParameterScope_Validate(t *testing.T) {
	config := ToolParameter{
		Name:     "test",
		Type:     TOOL_PARAMETER_TYPE_MODEL_SELECTOR,
		Scope:    parser.ToPtr("llm& document&tool-call"),
		Required: true,
		Label: I18nObject{
			ZhHans: "模型",
			EnUS:   "Model",
		},
		HumanDescription: I18nObject{
			ZhHans: "请选择模型",
			EnUS:   "Please select a model",
		},
		LLMDescription: "please select a model",
		Form:           TOOL_PARAMETER_FORM_FORM,
	}

	data := parser.MarshalJsonBytes(config)

	if _, err := parser.UnmarshalJsonBytes[ToolParameter](data); err != nil {
		t.Errorf("ParameterScope_Validate() error = %v", err)
	}
}
