package plugin_entities

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/manifest_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
	"gopkg.in/yaml.v3"
)

// TriggerRuntime represents the runtime context for trigger execution
type TriggerRuntime struct {
	Credentials map[string]any `json:"credentials" yaml:"credentials"`
	SessionID   *string        `json:"session_id" yaml:"session_id"`
}

// TriggerParameterOption represents the option of the trigger parameter
type TriggerParameterOption struct {
	Label I18nObject `json:"label" yaml:"label" validate:"required"`
	Value any        `json:"value" yaml:"value" validate:"required"`
}

// TriggerParameterType represents the type of the parameter
type TriggerParameterType string

const (
	TRIGGER_PARAMETER_TYPE_STRING         TriggerParameterType = STRING
	TRIGGER_PARAMETER_TYPE_NUMBER         TriggerParameterType = NUMBER
	TRIGGER_PARAMETER_TYPE_BOOLEAN        TriggerParameterType = BOOLEAN
	TRIGGER_PARAMETER_TYPE_SELECT         TriggerParameterType = SELECT
	TRIGGER_PARAMETER_TYPE_FILE           TriggerParameterType = FILE
	TRIGGER_PARAMETER_TYPE_FILES          TriggerParameterType = FILES
	TRIGGER_PARAMETER_TYPE_MODEL_SELECTOR TriggerParameterType = MODEL_SELECTOR
	TRIGGER_PARAMETER_TYPE_APP_SELECTOR   TriggerParameterType = APP_SELECTOR
	TRIGGER_PARAMETER_TYPE_OBJECT         TriggerParameterType = OBJECT
	TRIGGER_PARAMETER_TYPE_ARRAY          TriggerParameterType = ARRAY
	TRIGGER_PARAMETER_TYPE_DYNAMIC_SELECT TriggerParameterType = DYNAMIC_SELECT
	TRIGGER_PARAMETER_TYPE_CHECKBOX       TriggerParameterType = CHECKBOX
)

func isTriggerParameterType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case string(TRIGGER_PARAMETER_TYPE_STRING),
		string(TRIGGER_PARAMETER_TYPE_NUMBER),
		string(TRIGGER_PARAMETER_TYPE_BOOLEAN),
		string(TRIGGER_PARAMETER_TYPE_SELECT),
		string(TRIGGER_PARAMETER_TYPE_FILE),
		string(TRIGGER_PARAMETER_TYPE_FILES),
		string(TRIGGER_PARAMETER_TYPE_MODEL_SELECTOR),
		string(TRIGGER_PARAMETER_TYPE_APP_SELECTOR),
		string(TRIGGER_PARAMETER_TYPE_OBJECT),
		string(TRIGGER_PARAMETER_TYPE_ARRAY),
		string(TRIGGER_PARAMETER_TYPE_DYNAMIC_SELECT),
		string(TRIGGER_PARAMETER_TYPE_CHECKBOX):
		return true
	}
	return false
}

func init() {
	validators.GlobalEntitiesValidator.RegisterValidation("trigger_parameter_type", isTriggerParameterType)
}

// TriggerParameter represents the parameter of the trigger
type TriggerParameter struct {
	Name         string                   `json:"name" yaml:"name" validate:"required"`
	Label        I18nObject               `json:"label" yaml:"label" validate:"required"`
	Type         TriggerParameterType     `json:"type" yaml:"type" validate:"required,trigger_parameter_type"`
	AutoGenerate *ParameterAutoGenerate   `json:"auto_generate,omitempty" yaml:"auto_generate,omitempty"`
	Template     *ParameterTemplate       `json:"template,omitempty" yaml:"template,omitempty"`
	Scope        *string                  `json:"scope,omitempty" yaml:"scope,omitempty"`
	Required     bool                     `json:"required" yaml:"required"`
	Multiple     bool                     `json:"multiple,omitempty" yaml:"multiple,omitempty"`
	Default      any                      `json:"default,omitempty" yaml:"default,omitempty"`
	Min          *float64                 `json:"min,omitempty" yaml:"min,omitempty"`
	Max          *float64                 `json:"max,omitempty" yaml:"max,omitempty"`
	Precision    *int                     `json:"precision,omitempty" yaml:"precision,omitempty"`
	Options      []TriggerParameterOption `json:"options,omitempty" yaml:"options,omitempty" validate:"omitempty,dive"`
	Description  *I18nObject              `json:"description,omitempty" yaml:"description,omitempty"`
}

// TriggerProviderIdentity represents the identity of the trigger provider
type TriggerProviderIdentity struct {
	Author      string                        `json:"author" validate:"required"`
	Name        string                        `json:"name" validate:"required,tool_provider_identity_name"`
	Description I18nObject                    `json:"description"`
	Icon        string                        `json:"icon" validate:"required"`
	IconDark    string                        `json:"icon_dark" validate:"omitempty"`
	Label       I18nObject                    `json:"label" validate:"required"`
	Tags        []manifest_entities.PluginTag `json:"tags" validate:"omitempty,dive,plugin_tag"`
}

var triggerProviderIdentityNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func isTriggerProviderIdentityName(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return triggerProviderIdentityNameRegex.MatchString(value)
}

func init() {
	validators.GlobalEntitiesValidator.RegisterValidation("trigger_provider_identity_name", isTriggerProviderIdentityName)
}

// TriggerIdentity represents the identity of the trigger
type TriggerIdentity struct {
	Author string     `json:"author" yaml:"author" validate:"required"`
	Name   string     `json:"name" yaml:"name" validate:"required,trigger_identity_name"`
	Label  I18nObject `json:"label" yaml:"label" validate:"required"`
}

var triggerIdentityNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func isTriggerIdentityName(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return triggerIdentityNameRegex.MatchString(value)
}

func init() {
	validators.GlobalEntitiesValidator.RegisterValidation("trigger_identity_name", isTriggerIdentityName)
}

// TriggerDescription represents the description of the trigger
type TriggerDescription struct {
	Human I18nObject `json:"human" yaml:"human" validate:"required"`
	LLM   I18nObject `json:"llm" yaml:"llm" validate:"required"`
}

// TriggerDeclaration represents the configuration of a trigger
type TriggerDeclaration struct {
	Identity     TriggerIdentity    `json:"identity" yaml:"identity" validate:"required"`
	Parameters   []TriggerParameter `json:"parameters" yaml:"parameters" validate:"omitempty,dive"`
	Description  TriggerDescription `json:"description" yaml:"description" validate:"required"`
	OutputSchema map[string]any     `json:"output_schema,omitempty" yaml:"output_schema,omitempty"`
}

// SubscriptionConstructor represents the subscription constructor of the trigger provider
type SubscriptionConstructor struct {
	Parameters        []TriggerParameter `json:"parameters" yaml:"parameters" validate:"omitempty,dive"`
	CredentialsSchema []ProviderConfig   `json:"credentials_schema" yaml:"credentials_schema" validate:"omitempty,dive"`
	OAuthSchema       *OAuthSchema       `json:"oauth_schema,omitempty" yaml:"oauth_schema,omitempty" validate:"omitempty"`
}

// TriggerProviderDeclaration represents the configuration of a trigger provider
type TriggerProviderDeclaration struct {
	Identity                TriggerProviderIdentity  `json:"identity" yaml:"identity" validate:"required"`
	SubscriptionSchema      []ProviderConfig         `json:"subscription_schema" yaml:"subscription_schema" validate:"required"`
	SubscriptionConstructor *SubscriptionConstructor `json:"subscription_constructor" yaml:"subscription_constructor" validate:"omitempty"`
	Triggers                []TriggerDeclaration     `json:"triggers" yaml:"triggers" validate:"omitempty,dive"`
	TriggerFiles            []string                 `json:"-" yaml:"-"`
}

// Subscription represents the result of a successful trigger subscription operation
type Subscription struct {
	ExpiresAt  int64          `json:"expires_at" yaml:"expires_at" validate:"required"`
	Endpoint   string         `json:"endpoint" yaml:"endpoint" validate:"required"`
	Parameters map[string]any `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Properties map[string]any `json:"properties" yaml:"properties" validate:"required"`
}

// Unsubscription represents the result of a trigger unsubscription operation
type Unsubscription struct {
	Success bool    `json:"success" yaml:"success" validate:"required"`
	Message *string `json:"message,omitempty" yaml:"message,omitempty"`
}

// MarshalJSON implements custom JSON marshalling for TriggerProviderConfiguration
func (t *TriggerProviderDeclaration) MarshalJSON() ([]byte, error) {
	type alias TriggerProviderDeclaration
	p := alias(*t)
	if p.SubscriptionSchema == nil {
		p.SubscriptionSchema = []ProviderConfig{}
	}
	if p.SubscriptionConstructor.Parameters == nil {
		p.SubscriptionConstructor.Parameters = []TriggerParameter{}
	}
	if p.SubscriptionConstructor.CredentialsSchema == nil {
		p.SubscriptionConstructor.CredentialsSchema = []ProviderConfig{}
	}
	if p.Triggers == nil {
		p.Triggers = []TriggerDeclaration{}
	}
	return json.Marshal(p)
}

// convertYAMLNodeToProviderConfigList converts a YAML node to []ProviderConfig
// It supports both array format and dict/map format where keys become the Name field
func convertYAMLNodeToProviderConfigList(node *yaml.Node) ([]ProviderConfig, error) {
	if node.Kind != yaml.MappingNode {
		// not a map, decode as array directly
		configs := make([]ProviderConfig, 0)
		if err := node.Decode(&configs); err != nil {
			return nil, err
		}
		return configs, nil
	}

	// handle map/dict format: convert to array with keys as Name field
	configs := make([]ProviderConfig, 0, len(node.Content)/2)
	var currentKey string
	for i, item := range node.Content {
		if i%2 == 0 {
			// even indices are keys
			if item.Kind != yaml.ScalarNode {
				return nil, fmt.Errorf("expected scalar key at position %d", i)
			}
			currentKey = item.Value
		} else {
			// odd indices are values
			if item.Kind != yaml.MappingNode {
				return nil, fmt.Errorf("expected mapping value for key %s", currentKey)
			}
			var config ProviderConfig
			if err := item.Decode(&config); err != nil {
				return nil, fmt.Errorf("failed to decode config for key %s: %w", currentKey, err)
			}
			config.Name = currentKey
			configs = append(configs, config)
		}
	}
	return configs, nil
}

// UnmarshalYAML implements custom YAML unmarshalling for TriggerProviderConfiguration
func (t *TriggerProviderDeclaration) UnmarshalYAML(value *yaml.Node) error {
	type alias struct {
		Identity                TriggerProviderIdentity `yaml:"identity"`
		SubscriptionSchema      yaml.Node               `yaml:"subscription_schema"`
		SubscriptionConstructor yaml.Node               `yaml:"subscription_constructor"`
		Triggers                yaml.Node               `yaml:"triggers"`
	}

	var temp alias

	err := value.Decode(&temp)
	if err != nil {
		return err
	}

	// apply identity
	t.Identity = temp.Identity

	// handle subscription_schema conversion from dict to list format
	subscriptionSchema, err := convertYAMLNodeToProviderConfigList(&temp.SubscriptionSchema)
	if err != nil {
		return fmt.Errorf("failed to parse subscription_schema: %w", err)
	}
	t.SubscriptionSchema = subscriptionSchema

	// handle subscription_constructor
	if temp.SubscriptionConstructor.Kind == yaml.MappingNode {
		// parse subscription_constructor with custom handling for credentials_schema
		type constructorAlias struct {
			Parameters        yaml.Node `yaml:"parameters"`
			CredentialsSchema yaml.Node `yaml:"credentials_schema"`
			OAuthSchema       yaml.Node `yaml:"oauth_schema"`
		}

		var constructorTemp constructorAlias
		if err := temp.SubscriptionConstructor.Decode(&constructorTemp); err != nil {
			return err
		}

		t.SubscriptionConstructor = &SubscriptionConstructor{}

		// decode parameters if present
		if constructorTemp.Parameters.Kind != 0 {
			if err := constructorTemp.Parameters.Decode(&t.SubscriptionConstructor.Parameters); err != nil {
				return fmt.Errorf("failed to parse subscription_constructor.parameters: %w", err)
			}
		}

		// handle credentials_schema conversion from dict to list format
		if constructorTemp.CredentialsSchema.Kind != 0 {
			credentialsSchema, err := convertYAMLNodeToProviderConfigList(&constructorTemp.CredentialsSchema)
			if err != nil {
				return fmt.Errorf("failed to parse subscription_constructor.credentials_schema: %w", err)
			}
			t.SubscriptionConstructor.CredentialsSchema = credentialsSchema
		}

		// decode oauth_schema if present
		if constructorTemp.OAuthSchema.Kind != 0 {
			var oauthSchema OAuthSchema
			if err := constructorTemp.OAuthSchema.Decode(&oauthSchema); err != nil {
				return fmt.Errorf("failed to parse subscription_constructor.oauth_schema: %w", err)
			}
			t.SubscriptionConstructor.OAuthSchema = &oauthSchema
		}
	}

	// initialize TriggerFiles
	if t.TriggerFiles == nil {
		t.TriggerFiles = []string{}
	}

	// unmarshal triggers - support both file paths and direct declarations
	if temp.Triggers.Kind == yaml.SequenceNode {
		for _, item := range temp.Triggers.Content {
			if item.Kind == yaml.ScalarNode {
				// It's a string (file path), add to TriggerFiles
				t.TriggerFiles = append(t.TriggerFiles, item.Value)
			} else if item.Kind == yaml.MappingNode {
				// It's an object (direct trigger declaration), parse and add to Triggers
				trigger := TriggerDeclaration{}
				if err := item.Decode(&trigger); err != nil {
					return err
				}
				t.Triggers = append(t.Triggers, trigger)
			}
		}
	}

	// initialize empty arrays if nil
	if t.SubscriptionSchema == nil {
		t.SubscriptionSchema = []ProviderConfig{}
	}

	if t.SubscriptionConstructor != nil {
		if t.SubscriptionConstructor.Parameters == nil {
			t.SubscriptionConstructor.Parameters = []TriggerParameter{}
		}

		if t.SubscriptionConstructor.CredentialsSchema == nil {
			t.SubscriptionConstructor.CredentialsSchema = []ProviderConfig{}
		}
	}

	if t.Triggers == nil {
		t.Triggers = []TriggerDeclaration{}
	}

	if t.Identity.Tags == nil {
		t.Identity.Tags = []manifest_entities.PluginTag{}
	}

	return nil
}
