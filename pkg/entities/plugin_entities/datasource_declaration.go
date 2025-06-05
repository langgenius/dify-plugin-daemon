package plugin_entities

import (
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/manifest_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
	"gopkg.in/yaml.v3"
)

type DatasourceType string

const (
	DatasourceTypeWebsiteCrawl   DatasourceType = "website_crawl"
	DatasourceTypeOnlineDocument DatasourceType = "online_document"
)

func isDatasourceProviderType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case string(DatasourceTypeWebsiteCrawl),
		string(DatasourceTypeOnlineDocument):
		return true
	}
	return false
}

func init() {
	validators.GlobalEntitiesValidator.RegisterValidation("datasource_provider_type", isDatasourceProviderType)
}

type DatasourceIdentity struct {
	Author string     `json:"author" yaml:"author" validate:"required"`
	Name   string     `json:"name" yaml:"name" validate:"required"`
	Label  I18nObject `json:"label" yaml:"label" validate:"required"`
	Icon   string     `json:"icon" yaml:"icon" validate:"omitempty"`
}

type DatasourceParameterType string

const (
	DATASOURCE_PARAMETER_TYPE_STRING       DatasourceParameterType = STRING
	DATASOURCE_PARAMETER_TYPE_NUMBER       DatasourceParameterType = NUMBER
	DATASOURCE_PARAMETER_TYPE_BOOLEAN      DatasourceParameterType = BOOLEAN
	DATASOURCE_PARAMETER_TYPE_SELECT       DatasourceParameterType = SELECT
	DATASOURCE_PARAMETER_TYPE_SECRET_INPUT DatasourceParameterType = SECRET_INPUT
)

func isDatasourceParameterType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case string(DATASOURCE_PARAMETER_TYPE_STRING),
		string(DATASOURCE_PARAMETER_TYPE_NUMBER),
		string(DATASOURCE_PARAMETER_TYPE_BOOLEAN),
		string(DATASOURCE_PARAMETER_TYPE_SELECT),
		string(DATASOURCE_PARAMETER_TYPE_SECRET_INPUT):
		return true
	}
	return false
}

func init() {
	validators.GlobalEntitiesValidator.RegisterValidation("datasource_parameter_type", isDatasourceParameterType)
}

type DatasourceParameter struct {
	Name         string                  `json:"name" yaml:"name" validate:"required,gt=0,lt=1024"`
	Label        I18nObject              `json:"label" yaml:"label" validate:"required"`
	Type         DatasourceParameterType `json:"type" yaml:"type" validate:"required,datasource_parameter_type"`
	Scope        *string                 `json:"scope" yaml:"scope" validate:"omitempty,max=1024,is_scope"`
	Required     bool                    `json:"required" yaml:"required"`
	AutoGenerate *ParameterAutoGenerate  `json:"auto_generate" yaml:"auto_generate" validate:"omitempty"`
	Template     *ParameterTemplate      `json:"template" yaml:"template" validate:"omitempty"`
	Default      any                     `json:"default" yaml:"default" validate:"omitempty,is_basic_type"`
	Min          *float64                `json:"min" yaml:"min" validate:"omitempty"`
	Max          *float64                `json:"max" yaml:"max" validate:"omitempty"`
	Precision    *int                    `json:"precision" yaml:"precision" validate:"omitempty"`
	Options      []ParameterOption       `json:"options" yaml:"options" validate:"omitempty,dive"`
	Description  I18nObject              `json:"description" yaml:"description" validate:"required"`
}

type DatasourceDeclaration struct {
	Identity    DatasourceIdentity    `json:"identity" yaml:"identity" validate:"required"`
	Parameters  []DatasourceParameter `json:"parameters" yaml:"parameters" validate:"required,dive"`
	Description I18nObject            `json:"description" yaml:"description" validate:"required"`
}

type DatasourceProviderIdentity struct {
	Author      string                        `json:"author" yaml:"author" validate:"required"`
	Name        string                        `json:"name" yaml:"name" validate:"required"`
	Description I18nObject                    `json:"description" yaml:"description" validate:"required"`
	Icon        string                        `json:"icon" yaml:"icon" validate:"required"`
	Label       I18nObject                    `json:"label" yaml:"label" validate:"required"`
	Tags        []manifest_entities.PluginTag `json:"tags" yaml:"tags" validate:"omitempty,dive,plugin_tag"`
}

type DatasourceProviderDeclaration struct {
	Identity          DatasourceProviderIdentity `json:"identity" yaml:"identity" validate:"required"`
	CredentialsSchema []ProviderConfig           `json:"credentials_schema" yaml:"credentials_schema" validate:"omitempty,dive"`
	OAuthSchema       *OAuthSchema               `json:"oauth_schema" yaml:"oauth_schema" validate:"omitempty"`
	ProviderType      DatasourceType             `json:"provider_type" yaml:"provider_type" validate:"required,datasource_provider_type"`
	Datasources       []DatasourceDeclaration    `json:"datasources" yaml:"datasources" validate:"required,dive"`
	DatasourceFiles   []string                   `json:"-" yaml:"-"`
}

func (t *DatasourceProviderDeclaration) MarshalJSON() ([]byte, error) {
	type alias DatasourceProviderDeclaration
	p := alias(*t)
	if p.CredentialsSchema == nil {
		p.CredentialsSchema = []ProviderConfig{}
	}
	if p.Datasources == nil {
		p.Datasources = []DatasourceDeclaration{}
	}
	return json.Marshal(p)
}

func (t *DatasourceProviderDeclaration) UnmarshalYAML(value *yaml.Node) error {
	type alias struct {
		Identity          DatasourceProviderIdentity `yaml:"identity"`
		CredentialsSchema yaml.Node                  `yaml:"credentials_schema"`
		Datasources       yaml.Node                  `yaml:"datasources"`
		OAuthSchema       *OAuthSchema               `yaml:"oauth_schema"`
		ProviderType      DatasourceType             `yaml:"provider_type"`
	}

	var temp alias

	err := value.Decode(&temp)
	if err != nil {
		return err
	}

	// apply identity
	t.Identity = temp.Identity

	// apply oauth_schema
	t.OAuthSchema = temp.OAuthSchema

	// apply provider_type
	t.ProviderType = temp.ProviderType

	// check if credentials_schema is a map
	if temp.CredentialsSchema.Kind != yaml.MappingNode {
		// not a map, convert it into array
		credentialsSchema := make([]ProviderConfig, 0)
		if err := temp.CredentialsSchema.Decode(&credentialsSchema); err != nil {
			return err
		}
		t.CredentialsSchema = credentialsSchema
	} else if temp.CredentialsSchema.Kind == yaml.MappingNode {
		credentialsSchema := make([]ProviderConfig, 0, len(temp.CredentialsSchema.Content)/2)
		currentKey := ""
		currentValue := &ProviderConfig{}
		for _, item := range temp.CredentialsSchema.Content {
			if item.Kind == yaml.ScalarNode {
				currentKey = item.Value
			} else if item.Kind == yaml.MappingNode {
				currentValue = &ProviderConfig{}
				if err := item.Decode(currentValue); err != nil {
					return err
				}
				currentValue.Name = currentKey
				credentialsSchema = append(credentialsSchema, *currentValue)
			}
		}
		t.CredentialsSchema = credentialsSchema
	}

	if t.DatasourceFiles == nil {
		t.DatasourceFiles = []string{}
	}

	// unmarshal datasources
	if temp.Datasources.Kind == yaml.SequenceNode {
		for _, item := range temp.Datasources.Content {
			if item.Kind == yaml.ScalarNode {
				t.DatasourceFiles = append(t.DatasourceFiles, item.Value)
			} else if item.Kind == yaml.MappingNode {
				datasource := DatasourceDeclaration{}
				if err := item.Decode(&datasource); err != nil {
					return err
				}
				t.Datasources = append(t.Datasources, datasource)
			}
		}
	}

	if t.CredentialsSchema == nil {
		t.CredentialsSchema = []ProviderConfig{}
	}

	if t.Datasources == nil {
		t.Datasources = []DatasourceDeclaration{}
	}

	if t.Identity.Tags == nil {
		t.Identity.Tags = []manifest_entities.PluginTag{}
	}

	return nil
}

func (t *DatasourceProviderDeclaration) UnmarshalJSON(data []byte) error {
	type alias DatasourceProviderDeclaration

	var temp struct {
		alias
		CredentialsSchema json.RawMessage   `json:"credentials_schema"`
		Datasources       []json.RawMessage `json:"datasources"`
		OAuthSchema       *OAuthSchema      `json:"oauth_schema"`
		ProviderType      DatasourceType    `json:"provider_type"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	*t = DatasourceProviderDeclaration(temp.alias)

	// Determine the type of CredentialsSchema
	var raw_message map[string]json.RawMessage
	if err := json.Unmarshal(temp.CredentialsSchema, &raw_message); err == nil {
		// It's an object
		credentialsSchemaObject := make(map[string]ProviderConfig)
		if err := json.Unmarshal(temp.CredentialsSchema, &credentialsSchemaObject); err != nil {
			return fmt.Errorf("failed to unmarshal credentials_schema as object: %v", err)
		}
		for _, value := range credentialsSchemaObject {
			t.CredentialsSchema = append(t.CredentialsSchema, value)
		}
	} else {
		// It's likely an array
		var credentials_schema_array []ProviderConfig
		if err := json.Unmarshal(temp.CredentialsSchema, &credentials_schema_array); err != nil {
			return fmt.Errorf("failed to unmarshal credentials_schema as array: %v", err)
		}
		t.CredentialsSchema = credentials_schema_array
	}

	if t.DatasourceFiles == nil {
		t.DatasourceFiles = []string{}
	}

	// unmarshal tools
	for _, item := range temp.Datasources {
		datasource := DatasourceDeclaration{}
		if err := json.Unmarshal(item, &datasource); err != nil {
			// try to unmarshal it as a string directly
			t.DatasourceFiles = append(t.DatasourceFiles, string(item))
		} else {
			t.Datasources = append(t.Datasources, datasource)
		}
	}

	if t.CredentialsSchema == nil {
		t.CredentialsSchema = []ProviderConfig{}
	}

	if t.Datasources == nil {
		t.Datasources = []DatasourceDeclaration{}
	}

	if t.Identity.Tags == nil {
		t.Identity.Tags = []manifest_entities.PluginTag{}
	}

	return nil
}
