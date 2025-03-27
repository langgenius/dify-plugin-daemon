package plugin_entities

import (
	"encoding/json"

	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/manifest_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
	"gopkg.in/yaml.v3"
)

type AgentStrategyProviderIdentity struct {
	ToolProviderIdentity `json:",inline" yaml:",inline"`
}

type AgentStrategyIdentity struct {
	ToolIdentity `json:",inline" yaml:",inline"`
}

type AgentStrategyParameterType string

const (
	AGENT_STRATEGY_PARAMETER_TYPE_STRING         AgentStrategyParameterType = STRING
	AGENT_STRATEGY_PARAMETER_TYPE_NUMBER         AgentStrategyParameterType = NUMBER
	AGENT_STRATEGY_PARAMETER_TYPE_BOOLEAN        AgentStrategyParameterType = BOOLEAN
	AGENT_STRATEGY_PARAMETER_TYPE_SELECT         AgentStrategyParameterType = SELECT
	AGENT_STRATEGY_PARAMETER_TYPE_SECRET_INPUT   AgentStrategyParameterType = SECRET_INPUT
	AGENT_STRATEGY_PARAMETER_TYPE_FILE           AgentStrategyParameterType = FILE
	AGENT_STRATEGY_PARAMETER_TYPE_FILES          AgentStrategyParameterType = FILES
	AGENT_STRATEGY_PARAMETER_TYPE_APP_SELECTOR   AgentStrategyParameterType = APP_SELECTOR
	AGENT_STRATEGY_PARAMETER_TYPE_MODEL_SELECTOR AgentStrategyParameterType = MODEL_SELECTOR
	AGENT_STRATEGY_PARAMETER_TYPE_TOOLS_SELECTOR AgentStrategyParameterType = TOOLS_SELECTOR
)

func isAgentStrategyParameterType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case string(AGENT_STRATEGY_PARAMETER_TYPE_STRING),
		string(AGENT_STRATEGY_PARAMETER_TYPE_NUMBER),
		string(AGENT_STRATEGY_PARAMETER_TYPE_BOOLEAN),
		string(AGENT_STRATEGY_PARAMETER_TYPE_SELECT),
		string(AGENT_STRATEGY_PARAMETER_TYPE_SECRET_INPUT),
		string(AGENT_STRATEGY_PARAMETER_TYPE_FILE),
		string(AGENT_STRATEGY_PARAMETER_TYPE_FILES),
		// string(TOOL_PARAMETER_TYPE_TOOL_SELECTOR),
		string(AGENT_STRATEGY_PARAMETER_TYPE_APP_SELECTOR),
		string(AGENT_STRATEGY_PARAMETER_TYPE_MODEL_SELECTOR),
		string(AGENT_STRATEGY_PARAMETER_TYPE_TOOLS_SELECTOR):
		return true
	}
	return false
}

func init() {
	validators.GlobalEntitiesValidator.RegisterValidation("agent_strategy_parameter_type", isAgentStrategyParameterType)
}

type AgentStrategyParameter struct {
	Name         string                     `json:"name" yaml:"name" validate:"required,gt=0,lt=1024"`
	Label        I18nObject                 `json:"label" yaml:"label" validate:"required"`
	Type         AgentStrategyParameterType `json:"type" yaml:"type" validate:"required,agent_strategy_parameter_type"`
	AutoGenerate *ParameterAutoGenerate     `json:"auto_generate" yaml:"auto_generate" validate:"omitempty"`
	Template     *ParameterTemplate         `json:"template" yaml:"template" validate:"omitempty"`
	Scope        *string                    `json:"scope" yaml:"scope" validate:"omitempty,max=1024,is_scope"`
	Required     bool                       `json:"required" yaml:"required"`
	Default      any                        `json:"default" yaml:"default" validate:"omitempty,is_basic_type"`
	Min          *float64                   `json:"min" yaml:"min" validate:"omitempty"`
	Max          *float64                   `json:"max" yaml:"max" validate:"omitempty"`
	Precision    *int                       `json:"precision" yaml:"precision" validate:"omitempty"`
	Options      []ToolParameterOption      `json:"options" yaml:"options" validate:"omitempty,dive"`
}

type AgentStrategyOutputSchema map[string]any

type AgentStrategyDeclaration struct {
	Identity     AgentStrategyIdentity     `json:"identity" yaml:"identity" validate:"required"`
	Description  I18nObject                `json:"description" yaml:"description" validate:"required"`
	Parameters   []AgentStrategyParameter  `json:"parameters" yaml:"parameters" validate:"omitempty,dive"`
	OutputSchema AgentStrategyOutputSchema `json:"output_schema" yaml:"output_schema" validate:"omitempty,json_schema"`
	Features     []string                  `json:"features" yaml:"features" validate:"omitempty,dive,lt=256"`
}

type AgentStrategyProviderDeclaration struct {
	Identity      AgentStrategyProviderIdentity `json:"identity" yaml:"identity" validate:"required"`
	Strategies    []AgentStrategyDeclaration    `json:"strategies" yaml:"strategies" validate:"required,dive"`
	StrategyFiles []string                      `json:"-" yaml:"-"`
}

func (a *AgentStrategyProviderDeclaration) MarshalJSON() ([]byte, error) {
	type alias AgentStrategyProviderDeclaration
	p := alias(*a)
	if p.Strategies == nil {
		p.Strategies = []AgentStrategyDeclaration{}
	}

	for i := range p.Strategies {
		if p.Strategies[i].Features == nil {
			p.Strategies[i].Features = []string{}
		}
	}

	return json.Marshal(p)
}

func (a *AgentStrategyProviderDeclaration) UnmarshalYAML(value *yaml.Node) error {
	type alias struct {
		Identity   AgentStrategyProviderIdentity `yaml:"identity"`
		Strategies yaml.Node                     `yaml:"strategies"`
	}

	var temp alias

	err := value.Decode(&temp)
	if err != nil {
		return err
	}

	// apply identity
	a.Identity = temp.Identity

	if a.StrategyFiles == nil {
		a.StrategyFiles = []string{}
	}

	// unmarshal strategies
	if temp.Strategies.Kind == yaml.SequenceNode {
		for _, item := range temp.Strategies.Content {
			if item.Kind == yaml.ScalarNode {
				a.StrategyFiles = append(a.StrategyFiles, item.Value)
			} else if item.Kind == yaml.MappingNode {
				strategy := AgentStrategyDeclaration{}
				if err := item.Decode(&strategy); err != nil {
					return err
				}
				a.Strategies = append(a.Strategies, strategy)
			}
		}
	}

	if a.Strategies == nil {
		a.Strategies = []AgentStrategyDeclaration{}
	}

	if a.Identity.Tags == nil {
		a.Identity.Tags = []manifest_entities.PluginTag{}
	}

	return nil
}

func (a *AgentStrategyProviderDeclaration) UnmarshalJSON(data []byte) error {
	type alias AgentStrategyProviderDeclaration

	var temp struct {
		alias
		Strategies []json.RawMessage `json:"strategies"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	*a = AgentStrategyProviderDeclaration(temp.alias)

	// unmarshal strategies
	for _, item := range temp.Strategies {
		strategy := AgentStrategyDeclaration{}
		if err := json.Unmarshal(item, &strategy); err != nil {
			a.StrategyFiles = append(a.StrategyFiles, string(item))
		} else {
			a.Strategies = append(a.Strategies, strategy)
		}
	}

	if a.Strategies == nil {
		a.Strategies = []AgentStrategyDeclaration{}
	}

	if a.Identity.Tags == nil {
		a.Identity.Tags = []manifest_entities.PluginTag{}
	}

	return nil
}
