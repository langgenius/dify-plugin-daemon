package plugin_entities

import (
	"encoding/json"

	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
	"gopkg.in/yaml.v3"
)

type EndpointMethod string

const (
	EndpointMethodHead    EndpointMethod = "HEAD"
	EndpointMethodGet     EndpointMethod = "GET"
	EndpointMethodPost    EndpointMethod = "POST"
	EndpointMethodPut     EndpointMethod = "PUT"
	EndpointMethodDelete  EndpointMethod = "DELETE"
	EndpointMethodOptions EndpointMethod = "OPTIONS"
)

func isAvailableMethod(fl validator.FieldLevel) bool {
	method := fl.Field().String()
	switch method {
	case string(EndpointMethodHead),
		string(EndpointMethodGet),
		string(EndpointMethodPost),
		string(EndpointMethodPut),
		string(EndpointMethodDelete),
		string(EndpointMethodOptions):
		return true
	}
	return false
}

func init() {
	validators.GlobalEntitiesValidator.RegisterValidation("is_available_endpoint_method", isAvailableMethod)
}

type EndpointDeclaration struct {
	Path   string         `json:"path" yaml:"path" validate:"required"`
	Method EndpointMethod `json:"method" yaml:"method" validate:"required,is_available_endpoint_method"`
	Hidden bool           `json:"hidden" yaml:"hidden" validate:"omitempty"`
}

type EndpointProviderDeclaration struct {
	Settings      []ProviderConfig      `json:"settings" yaml:"settings" validate:"omitempty,dive"`
	Endpoints     []EndpointDeclaration `json:"endpoints" yaml:"endpoint_declarations" validate:"omitempty,dive"`
	EndpointFiles []string              `json:"-" yaml:"-"`
}

func (e *EndpointProviderDeclaration) UnmarshalYAML(node *yaml.Node) error {
	type alias EndpointProviderDeclaration

	var temp struct {
		alias     `yaml:",inline"`
		Endpoints yaml.Node `yaml:"endpoints"`
	}

	if err := node.Decode(&temp); err != nil {
		return err
	}

	e.Settings = temp.Settings

	if temp.Endpoints.Kind == yaml.SequenceNode {
		for _, node := range temp.Endpoints.Content {
			if node.Kind == yaml.ScalarNode {
				e.EndpointFiles = append(e.EndpointFiles, node.Value)
			} else {
				var declaration EndpointDeclaration
				if err := node.Decode(&declaration); err != nil {
					return nil
				}
				e.Endpoints = append(e.Endpoints, declaration)
			}
		}
	}

	return nil
}

func (e *EndpointProviderDeclaration) UnmarshalJSON(data []byte) error {
	type alias EndpointProviderDeclaration

	var temp struct {
		alias
		Endpoints json.RawMessage `json:"endpoints"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	*e = EndpointProviderDeclaration(temp.alias)

	if len(temp.Endpoints) == 0 {
		return nil
	}

	var raw_endpoints []json.RawMessage
	if err := json.Unmarshal(temp.Endpoints, &raw_endpoints); err != nil {
		return err
	}

	for _, raw_endpoint := range raw_endpoints {
		var declaration EndpointDeclaration
		if err := json.Unmarshal(raw_endpoint, &declaration); err != nil {
			e.EndpointFiles = append(e.EndpointFiles, string(raw_endpoint))
		} else {
			e.Endpoints = append(e.Endpoints, declaration)
		}
	}

	return nil
}
