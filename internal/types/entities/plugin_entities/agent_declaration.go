package plugin_entities

import (
	"encoding/json"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/manifest_entities"
	"gopkg.in/yaml.v3"
)

type AgentStrategyProviderIdentity struct {
	ToolProviderIdentity `json:",inline" yaml:",inline"`
}

type AgentStrategyIdentity struct {
	ToolIdentity `json:",inline" yaml:",inline"`
}

type AgentStrategyParameter struct {
	ToolParameter `json:",inline" yaml:",inline"`
}

type AgentStrategyOutputSchema map[string]any

type AgentStrategyDeclaration struct {
	Identity     AgentStrategyIdentity     `json:"identity" yaml:"identity" validate:"required"`
	Description  I18nObject                `json:"description" yaml:"description" validate:"required"`
	Parameters   []AgentStrategyParameter  `json:"parameters" yaml:"parameters" validate:"omitempty,dive"`
	OutputSchema AgentStrategyOutputSchema `json:"output_schema" yaml:"output_schema" validate:"omitempty,json_schema"`
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
