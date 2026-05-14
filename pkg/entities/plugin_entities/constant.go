package plugin_entities

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

const (
	SECRET_INPUT   = "secret-input"
	TEXT_INPUT     = "text-input"
	SELECT         = "select"
	STRING         = "string"
	NUMBER         = "number"
	FILE           = "file"
	FILES          = "files"
	BOOLEAN        = "boolean"
	APP_SELECTOR   = "app-selector"
	MODEL_SELECTOR = "model-selector"
	// TOOL_SELECTOR  = "tool-selector"
	TOOLS_SELECTOR = "array[tools]"
	ANY            = "any"
	// DynamicSelect
	DYNAMIC_SELECT = "dynamic-select"
	ARRAY          = "array"
	OBJECT         = "object"
	CHECKBOX       = "checkbox"
)

type ParameterOption struct {
	Value  string                      `json:"value" yaml:"value" validate:"required"`
	Label  I18nObject                  `json:"label" yaml:"label" validate:"required"`
	Icon   string                      `json:"icon" yaml:"icon" validate:"omitempty"`
	ShowOn []ToolParameterShowOnObject `json:"show_on" yaml:"show_on" validate:"omitempty,lte=16,dive"`
}

func (p *ParameterOption) UnmarshalJSON(data []byte) error {
	type Alias ParameterOption
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	if p.ShowOn == nil {
		p.ShowOn = []ToolParameterShowOnObject{}
	}

	return nil
}

func (p *ParameterOption) UnmarshalYAML(value *yaml.Node) error {
	type Alias ParameterOption
	aux := &struct {
		*Alias `yaml:",inline"`
	}{
		Alias: (*Alias)(p),
	}

	if err := value.Decode(&aux); err != nil {
		return err
	}

	if p.ShowOn == nil {
		p.ShowOn = []ToolParameterShowOnObject{}
	}

	return nil
}
