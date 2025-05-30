package plugin_entities

import (
	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/manifest_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
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
