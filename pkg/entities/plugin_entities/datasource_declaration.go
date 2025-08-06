package plugin_entities

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/manifest_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

type DatasourceType string

const (
	DatasourceTypeWebsiteCrawl   DatasourceType = "website_crawl"
	DatasourceTypeOnlineDocument DatasourceType = "online_document"
	DatasourceTypeOnlineDrive    DatasourceType = "online_drive"
)

func isDatasourceProviderType(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case string(DatasourceTypeWebsiteCrawl),
		string(DatasourceTypeOnlineDocument),
		string(DatasourceTypeOnlineDrive):
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

func isDatasourceOutputSchema(fl validator.FieldLevel) bool {
	// get schema from interface
	schemaMapInf := fl.Field().Interface()
	// convert to map[string]any
	datasourceSchemaMap, ok := schemaMapInf.(DatasourceOutputSchema)
	if !ok {
		return false
	}
	_, err := gojsonschema.NewSchema(gojsonschema.NewGoLoader(datasourceSchemaMap))
	if err != nil {
		return false
	}

	return err == nil
}

func init() {
	validators.GlobalEntitiesValidator.RegisterValidation("datasource_parameter_type", isDatasourceParameterType)
	validators.GlobalEntitiesValidator.RegisterValidation("datasource_output_schema", isDatasourceOutputSchema)
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

type DatasourceOutputSchema map[string]any

type DatasourceDeclaration struct {
	Identity     DatasourceIdentity     `json:"identity" yaml:"identity" validate:"required"`
	Parameters   []DatasourceParameter  `json:"parameters" yaml:"parameters" validate:"required,dive"`
	Description  I18nObject             `json:"description" yaml:"description" validate:"required"`
	OutputSchema DatasourceOutputSchema `json:"output_schema" yaml:"output_schema" validate:"omitempty,datasource_output_schema"`
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
		Datasources []json.RawMessage `json:"datasources"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	*t = DatasourceProviderDeclaration(temp.alias)

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

var BuiltinDefinitions = map[string]any{
	"file": map[string]any{
		"type": "object",
		"properties": map[string]any{
			"dify_builtin_type": map[string]any{
				"type":        "string",
				"enum":        []string{"File"},
				"description": "Business type identifier for frontend",
			},
			"name": map[string]any{
				"type":        "string",
				"description": "file name",
			},
			"size": map[string]any{
				"type":        "number",
				"description": "file size",
			},
			"file_type": map[string]any{
				"type":        "string",
				"description": "file type",
			},
			"extension": map[string]any{
				"type":        "string",
				"description": "file extension",
			},
			"mime_type": map[string]any{
				"type":        "string",
				"description": "file mime type",
			},
			"transfer_method": map[string]any{
				"type":        "string",
				"description": "file transfer method",
			},
			"url": map[string]any{
				"type":        "string",
				"description": "file url",
			},
			"related_id": map[string]any{
				"type":        "string",
				"description": "file related id",
			},
		},
		"required": []string{"name"},
	},
	"website_crawl": map[string]any{
		"type": "object",
		"properties": map[string]any{
			"dify_builtin_type": map[string]any{
				"type":        "string",
				"enum":        []string{"WebsiteCrawl"},
				"description": "Business type identifier for frontend",
			},
			"source_url": map[string]any{
				"type":        "string",
				"description": "The URL of the crawled website",
			},
			"content": map[string]any{
				"type":        "string",
				"description": "The content of the crawled website",
			},
			"title": map[string]any{
				"type":        "string",
				"description": "The title of the crawled website",
			},
			"description": map[string]any{
				"type":        "string",
				"description": "The description of the crawled website",
			},
		},
		"required": []string{"source_url", "content"},
	},
	"online_document": map[string]any{
		"type": "object",
		"properties": map[string]any{
			"dify_builtin_type": map[string]any{
				"type":        "string",
				"enum":        []string{"OnlineDocument"},
				"description": "Business type identifier for frontend",
			},
			"workspace_id": map[string]any{
				"type":        "string",
				"description": "The ID of the workspace where the document is stored",
			},
			"page_id": map[string]any{
				"type":        "string",
				"description": "The ID of the page in the document",
			},
			"content": map[string]any{
				"type":        "string",
				"description": "The content of the online document",
			},
		},
		"required": []string{"content"},
	},
	"general_structure_chunk": map[string]any{
		"type": "object",
		"properties": map[string]any{
			"dify_builtin_type": map[string]any{
				"type":        "string",
				"enum":        []string{"GeneralStructureChunk"},
				"description": "Business type identifier for frontend",
			},
			"general_chunks": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
				"description": "List of general content chunks",
			},
		},
		"required": []string{"general_chunks"},
	},
	"parent_child_structure_chunk": map[string]any{
		"type": "object",
		"properties": map[string]any{
			"dify_builtin_type": map[string]any{
				"type":        "string",
				"enum":        []string{"ParentChildStructureChunk"},
				"description": "Business type identifier for frontend",
			},
			"parent_mode": map[string]any{
				"type":        "string",
				"description": "The mode of parent-child relationship",
			},
			"parent_child_chunks": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"parent_content": map[string]any{
							"type":        "string",
							"description": "The parent content",
						},
						"child_contents": map[string]any{
							"type": "array",
							"items": map[string]any{
								"type": "string",
							},
							"description": "List of child contents",
						},
					},
					"required": []string{"parent_content", "child_contents"},
				},
				"description": "List of parent-child chunk pairs",
			},
		},
		"required": []string{"parent_mode", "parent_child_chunks"},
	},
	"qa_structure_chunk": map[string]any{
		"type": "object",
		"properties": map[string]any{
			"dify_builtin_type": map[string]any{
				"type":        "string",
				"enum":        []string{"QAStructureChunk"},
				"description": "Business type identifier for frontend",
			},
			"qa_chunks": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"question": map[string]any{
							"type":        "string",
							"description": "The question",
						},
						"answer": map[string]any{
							"type":        "string",
							"description": "The answer",
						},
					},
					"required": []string{"question", "answer"},
				},
				"description": "List of question-answer pairs",
			},
		},
		"required": []string{"qa_chunks"},
	},
}

func ResolveSchemaRefs(schema any, definitions map[string]any) (any, error) {
	switch v := schema.(type) {
	case map[string]any:
		if ref, hasRef := v["$ref"].(string); hasRef {
			if strings.HasPrefix(ref, "#/$defs/") {
				typeName := strings.TrimPrefix(ref, "#/$defs/")
				if def, exists := definitions[typeName]; exists {
					return ResolveSchemaRefs(def, definitions)
				}
				return nil, fmt.Errorf("reference '%s' not found in definitions", ref)
			}
			return nil, fmt.Errorf("unsupported reference format: %s", ref)
		}

		resolved := make(map[string]any)
		for key, value := range v {
			resolvedValue, err := ResolveSchemaRefs(value, definitions)
			if err != nil {
				return nil, err
			}
			resolved[key] = resolvedValue
		}
		return resolved, nil

	case []any:
		resolved := make([]any, len(v))
		for i, item := range v {
			resolvedItem, err := ResolveSchemaRefs(item, definitions)
			if err != nil {
				return nil, err
			}
			resolved[i] = resolvedItem
		}
		return resolved, nil

	default:
		return schema, nil
	}
}

func ProcessDatasourceYAML(yamlData map[string]any) (map[string]any, error) {
	if outputSchema, hasOutputSchema := yamlData["output_schema"]; hasOutputSchema {
		userDefinitions := make(map[string]any)
		if defs, hasDefs := yamlData["definitions"].(map[string]any); hasDefs {
			userDefinitions = defs
		}

		allDefinitions := make(map[string]any)
		for k, v := range BuiltinDefinitions {
			allDefinitions[k] = v
		}
		for k, v := range userDefinitions {
			allDefinitions[k] = v
		}

		resolvedSchema, err := ResolveSchemaRefs(outputSchema, allDefinitions)
		if err != nil {
			return nil, errors.Join(err, errors.New("failed to resolve schema references"))
		}

		yamlData["output_schema"] = resolvedSchema

		delete(yamlData, "definitions")
	}

	return yamlData, nil
}
