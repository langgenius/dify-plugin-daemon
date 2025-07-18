package plugin_entities

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/constants"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/manifest_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
)

type PluginCategory string

const (
	PLUGIN_CATEGORY_TOOL           PluginCategory = "tool"
	PLUGIN_CATEGORY_MODEL          PluginCategory = "model"
	PLUGIN_CATEGORY_EXTENSION      PluginCategory = "extension"
	PLUGIN_CATEGORY_AGENT_STRATEGY PluginCategory = "agent-strategy"
	PLUGIN_CATEGORY_DATASOURCE     PluginCategory = "datasource"
)

type PluginPermissionRequirement struct {
	Tool     *PluginPermissionToolRequirement     `json:"tool,omitempty" yaml:"tool,omitempty" validate:"omitempty"`
	Model    *PluginPermissionModelRequirement    `json:"model,omitempty" yaml:"model,omitempty" validate:"omitempty"`
	Node     *PluginPermissionNodeRequirement     `json:"node,omitempty" yaml:"node,omitempty" validate:"omitempty"`
	Endpoint *PluginPermissionEndpointRequirement `json:"endpoint,omitempty" yaml:"endpoint,omitempty" validate:"omitempty"`
	App      *PluginPermissionAppRequirement      `json:"app,omitempty" yaml:"app,omitempty" validate:"omitempty"`
	Storage  *PluginPermissionStorageRequirement  `json:"storage,omitempty" yaml:"storage,omitempty" validate:"omitempty"`
}

func (p *PluginPermissionRequirement) AllowInvokeTool() bool {
	return p != nil && p.Tool != nil && p.Tool.Enabled
}

func (p *PluginPermissionRequirement) AllowInvokeModel() bool {
	return p != nil && p.Model != nil && p.Model.Enabled
}

func (p *PluginPermissionRequirement) AllowInvokeLLM() bool {
	return p != nil && p.Model != nil && p.Model.Enabled && p.Model.LLM
}

func (p *PluginPermissionRequirement) AllowInvokeTextEmbedding() bool {
	return p != nil && p.Model != nil && p.Model.Enabled && p.Model.TextEmbedding
}

func (p *PluginPermissionRequirement) AllowInvokeRerank() bool {
	return p != nil && p.Model != nil && p.Model.Enabled && p.Model.Rerank
}

func (p *PluginPermissionRequirement) AllowInvokeTTS() bool {
	return p != nil && p.Model != nil && p.Model.Enabled && p.Model.TTS
}

func (p *PluginPermissionRequirement) AllowInvokeSpeech2Text() bool {
	return p != nil && p.Model != nil && p.Model.Enabled && p.Model.Speech2text
}

func (p *PluginPermissionRequirement) AllowInvokeModeration() bool {
	return p != nil && p.Model != nil && p.Model.Enabled && p.Model.Moderation
}

func (p *PluginPermissionRequirement) AllowInvokeNode() bool {
	return p != nil && p.Node != nil && p.Node.Enabled
}

func (p *PluginPermissionRequirement) AllowInvokeApp() bool {
	return p != nil && p.App != nil && p.App.Enabled
}

func (p *PluginPermissionRequirement) AllowRegisterEndpoint() bool {
	return p != nil && p.Endpoint != nil && p.Endpoint.Enabled
}

func (p *PluginPermissionRequirement) AllowInvokeStorage() bool {
	return p != nil && p.Storage != nil && p.Storage.Enabled
}

type PluginPermissionToolRequirement struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}

type PluginPermissionModelRequirement struct {
	Enabled       bool `json:"enabled" yaml:"enabled"`
	LLM           bool `json:"llm" yaml:"llm"`
	TextEmbedding bool `json:"text_embedding" yaml:"text_embedding"`
	Rerank        bool `json:"rerank" yaml:"rerank"`
	TTS           bool `json:"tts" yaml:"tts"`
	Speech2text   bool `json:"speech2text" yaml:"speech2text"`
	Moderation    bool `json:"moderation" yaml:"moderation"`
}

type PluginPermissionNodeRequirement struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}

type PluginPermissionEndpointRequirement struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}

type PluginPermissionAppRequirement struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}

type PluginPermissionStorageRequirement struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	Size    uint64 `json:"size" yaml:"size" validate:"min=1024,max=1073741824"` // min 1024 bytes, max 1G
}

type PluginResourceRequirement struct {
	// Memory in bytes
	Memory int64 `json:"memory" yaml:"memory" validate:"required"`
	// Permission requirements
	Permission *PluginPermissionRequirement `json:"permission,omitempty" yaml:"permission,omitempty" validate:"omitempty"`
}

type PluginDeclarationPlatformArch string

type PluginRunner struct {
	Language   constants.Language `json:"language" yaml:"language" validate:"required,is_available_language"`
	Version    string             `json:"version" yaml:"version" validate:"required,max=128"`
	Entrypoint string             `json:"entrypoint" yaml:"entrypoint" validate:"required,max=256"`
}

type PluginMeta struct {
	Version            string           `json:"version" yaml:"version" validate:"required,version"`
	Arch               []constants.Arch `json:"arch" yaml:"arch" validate:"required,dive,is_available_arch"`
	Runner             PluginRunner     `json:"runner" yaml:"runner" validate:"required"`
	MinimumDifyVersion *string          `json:"minimum_dify_version" yaml:"minimum_dify_version"`
}

type PluginExtensions struct {
	Tools           []string `json:"tools" yaml:"tools,omitempty" validate:"omitempty,dive,max=128"`
	Models          []string `json:"models" yaml:"models,omitempty" validate:"omitempty,dive,max=128"`
	Endpoints       []string `json:"endpoints" yaml:"endpoints,omitempty" validate:"omitempty,dive,max=128"`
	AgentStrategies []string `json:"agent_strategies" yaml:"agent_strategies,omitempty" validate:"omitempty,dive,max=128"`
	Datasources     []string `json:"datasources" yaml:"datasources,omitempty" validate:"omitempty,dive,max=128"`
}

type PluginDeclarationWithoutAdvancedFields struct {
	Version     manifest_entities.Version          `json:"version" yaml:"version,omitempty" validate:"required,version"`
	Type        manifest_entities.DifyManifestType `json:"type" yaml:"type,omitempty" validate:"required,eq=plugin"`
	Author      string                             `json:"author" yaml:"author,omitempty" validate:"omitempty,max=64"`
	Name        string                             `json:"name" yaml:"name,omitempty" validate:"required,max=128"`
	Label       I18nObject                         `json:"label" yaml:"label" validate:"required"`
	Description I18nObject                         `json:"description" yaml:"description" validate:"required"`
	Icon        string                             `json:"icon" yaml:"icon,omitempty" validate:"required,max=128"`
	IconDark    string                             `json:"icon_dark" yaml:"icon_dark,omitempty" validate:"omitempty,max=128"`
	Resource    PluginResourceRequirement          `json:"resource" yaml:"resource,omitempty" validate:"required"`
	Plugins     PluginExtensions                   `json:"plugins" yaml:"plugins,omitempty" validate:"required"`
	Meta        PluginMeta                         `json:"meta" yaml:"meta,omitempty" validate:"required"`
	Tags        []manifest_entities.PluginTag      `json:"tags" yaml:"tags,omitempty" validate:"omitempty,dive,plugin_tag,max=128"`
	CreatedAt   time.Time                          `json:"created_at" yaml:"created_at,omitempty" validate:"required"`
	Privacy     *string                            `json:"privacy,omitempty" yaml:"privacy,omitempty" validate:"omitempty"`
	Repo        *string                            `json:"repo,omitempty" yaml:"repo,omitempty" validate:"omitempty,url"`
}

func (p *PluginDeclarationWithoutAdvancedFields) UnmarshalJSON(data []byte) error {
	type Alias PluginDeclarationWithoutAdvancedFields
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	if p.Tags == nil {
		p.Tags = []manifest_entities.PluginTag{}
	}

	return nil
}

type PluginDeclaration struct {
	PluginDeclarationWithoutAdvancedFields `yaml:",inline"`
	Verified                               bool                              `json:"verified" yaml:"verified"`
	Endpoint                               *EndpointProviderDeclaration      `json:"endpoint,omitempty" yaml:"endpoint,omitempty" validate:"omitempty"`
	Model                                  *ModelProviderDeclaration         `json:"model,omitempty" yaml:"model,omitempty" validate:"omitempty"`
	Tool                                   *ToolProviderDeclaration          `json:"tool,omitempty" yaml:"tool,omitempty" validate:"omitempty"`
	AgentStrategy                          *AgentStrategyProviderDeclaration `json:"agent_strategy,omitempty" yaml:"agent_strategy,omitempty" validate:"omitempty"`
	Datasource                             *DatasourceProviderDeclaration    `json:"datasource,omitempty" yaml:"datasource,omitempty" validate:"omitempty"`
}

func (p *PluginDeclaration) Category() PluginCategory {
	if p.Tool != nil || len(p.Plugins.Tools) != 0 {
		return PLUGIN_CATEGORY_TOOL
	}
	if p.Model != nil || len(p.Plugins.Models) != 0 {
		return PLUGIN_CATEGORY_MODEL
	}
	if p.Datasource != nil || len(p.Plugins.Datasources) != 0 {
		return PLUGIN_CATEGORY_DATASOURCE
	}
	if p.AgentStrategy != nil || len(p.Plugins.AgentStrategies) != 0 {
		return PLUGIN_CATEGORY_AGENT_STRATEGY
	}
	return PLUGIN_CATEGORY_EXTENSION
}

func (p *PluginDeclaration) UnmarshalJSON(data []byte) error {
	// First unmarshal the embedded struct
	if err := json.Unmarshal(data, &p.PluginDeclarationWithoutAdvancedFields); err != nil {
		return err
	}

	// Then unmarshal the remaining fields
	type PluginExtra struct {
		Verified      bool                              `json:"verified"`
		Endpoint      *EndpointProviderDeclaration      `json:"endpoint,omitempty"`
		Model         *ModelProviderDeclaration         `json:"model,omitempty"`
		Tool          *ToolProviderDeclaration          `json:"tool,omitempty"`
		AgentStrategy *AgentStrategyProviderDeclaration `json:"agent_strategy,omitempty"`
		Datasource    *DatasourceProviderDeclaration    `json:"datasource,omitempty"`
	}

	var extra PluginExtra
	if err := json.Unmarshal(data, &extra); err != nil {
		return err
	}

	p.Verified = extra.Verified
	p.Endpoint = extra.Endpoint
	p.Model = extra.Model
	p.Tool = extra.Tool
	p.AgentStrategy = extra.AgentStrategy
	p.Datasource = extra.Datasource

	return nil
}

func (p *PluginDeclaration) MarshalJSON() ([]byte, error) {
	// TODO: performance issue, need a better way to do this
	c := *p
	c.FillInDefaultValues()
	type alias PluginDeclaration
	return json.Marshal(alias(c))
}

var (
	PluginNameRegex = regexp.MustCompile(`^[a-z0-9_-]{1,128}$`)
	AuthorRegex     = regexp.MustCompile(`^[a-z0-9_-]{1,64}$`)
)

func isPluginName(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return PluginNameRegex.MatchString(value)
}

func (p *PluginDeclaration) Identity() string {
	return parser.MarshalPluginID(p.Author, p.Name, p.Version.String())
}

func (p *PluginDeclaration) ManifestValidate() error {
	if p.Endpoint == nil && p.Model == nil && p.Tool == nil && p.AgentStrategy == nil && p.Datasource == nil {
		return fmt.Errorf("at least one of endpoint, model, tool, agent_strategy, or datasource must be provided")
	}

	if p.Model != nil {
		if p.Datasource != nil || p.Tool != nil || p.Endpoint != nil || p.AgentStrategy != nil {
			return fmt.Errorf("model and datasource, tool, endpoint, or agent_strategy cannot be provided at the same time")
		}
	}

	if p.AgentStrategy != nil {
		if p.Tool != nil || p.Model != nil || p.Endpoint != nil || p.Datasource != nil {
			return fmt.Errorf("agent_strategy and tool, model, endpoint, or datasource cannot be provided at the same time")
		}
	}

	if p.Datasource != nil {
		if p.Tool != nil || p.Model != nil || p.Endpoint != nil || p.AgentStrategy != nil {
			return fmt.Errorf("datasource and tool, model, endpoint, or agent_strategy cannot be provided at the same time")
		}
	}

	return nil
}

func (p *PluginDeclaration) FillInDefaultValues() {
	if p.Tool != nil {
		if p.Tool.Identity.Description.EnUS == "" {
			p.Tool.Identity.Description = p.Description
		}

		if len(p.Tool.Identity.Tags) == 0 {
			p.Tool.Identity.Tags = p.Tags
		}
	}

	if p.Model != nil {
		if p.Model.Description == nil {
			deepCopiedDescription := p.Description
			p.Model.Description = &deepCopiedDescription
		}
	}

	if p.Tags == nil {
		p.Tags = []manifest_entities.PluginTag{}
	}
}

func init() {
	// init validator
	validators.GlobalEntitiesValidator.RegisterValidation("plugin_name", isPluginName)
}

func UnmarshalPluginDeclarationFromYaml(data []byte) (*PluginDeclaration, error) {
	obj, err := parser.UnmarshalYamlBytes[PluginDeclaration](data)
	if err != nil {
		return nil, err
	}

	if err := validators.GlobalEntitiesValidator.Struct(obj); err != nil {
		return nil, err
	}

	obj.FillInDefaultValues()

	return &obj, nil
}

func UnmarshalPluginDeclarationFromJSON(data []byte) (*PluginDeclaration, error) {
	obj, err := parser.UnmarshalJsonBytes[PluginDeclaration](data)
	if err != nil {
		return nil, err
	}

	if err := validators.GlobalEntitiesValidator.Struct(obj); err != nil {
		return nil, err
	}

	obj.FillInDefaultValues()

	return &obj, nil
}
