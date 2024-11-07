package plugin_entities

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/constants"
	"github.com/langgenius/dify-plugin-daemon/internal/types/validators"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

type DifyManifestType string

const (
	PluginType DifyManifestType = "plugin"
)

type PluginTag string

const (
	PLUGIN_TAG_SEARCH        PluginTag = "search"
	PLUGIN_TAG_IMAGE         PluginTag = "image"
	PLUGIN_TAG_VIDEOS        PluginTag = "videos"
	PLUGIN_TAG_WEATHER       PluginTag = "weather"
	PLUGIN_TAG_FINANCE       PluginTag = "finance"
	PLUGIN_TAG_DESIGN        PluginTag = "design"
	PLUGIN_TAG_TRAVEL        PluginTag = "travel"
	PLUGIN_TAG_SOCIAL        PluginTag = "social"
	PLUGIN_TAG_NEWS          PluginTag = "news"
	PLUGIN_TAG_MEDICAL       PluginTag = "medical"
	PLUGIN_TAG_PRODUCTIVITY  PluginTag = "productivity"
	PLUGIN_TAG_EDUCATION     PluginTag = "education"
	PLUGIN_TAG_BUSINESS      PluginTag = "business"
	PLUGIN_TAG_ENTERTAINMENT PluginTag = "entertainment"
	PLUGIN_TAG_UTILITIES     PluginTag = "utilities"
	PLUGIN_TAG_OTHER         PluginTag = "other"
)

func isPluginTag(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case string(PLUGIN_TAG_SEARCH),
		string(PLUGIN_TAG_IMAGE),
		string(PLUGIN_TAG_VIDEOS),
		string(PLUGIN_TAG_WEATHER),
		string(PLUGIN_TAG_FINANCE),
		string(PLUGIN_TAG_DESIGN),
		string(PLUGIN_TAG_TRAVEL),
		string(PLUGIN_TAG_SOCIAL),
		string(PLUGIN_TAG_NEWS),
		string(PLUGIN_TAG_MEDICAL),
		string(PLUGIN_TAG_PRODUCTIVITY),
		string(PLUGIN_TAG_EDUCATION),
		string(PLUGIN_TAG_BUSINESS),
		string(PLUGIN_TAG_ENTERTAINMENT),
		string(PLUGIN_TAG_UTILITIES),
		string(PLUGIN_TAG_OTHER):
		return true
	}
	return false
}

func init() {
	validators.GlobalEntitiesValidator.RegisterValidation("plugin_tag", isPluginTag)
}

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
	Version string           `json:"version" yaml:"version" validate:"required,version"`
	Arch    []constants.Arch `json:"arch" yaml:"arch" validate:"required,dive,is_available_arch"`
	Runner  PluginRunner     `json:"runner" yaml:"runner" validate:"required"`
}

type PluginExtensions struct {
	Tools     []string `json:"tools" yaml:"tools,omitempty" validate:"omitempty,dive,max=128"`
	Models    []string `json:"models" yaml:"models,omitempty" validate:"omitempty,dive,max=128"`
	Endpoints []string `json:"endpoints" yaml:"endpoints,omitempty" validate:"omitempty,dive,max=128"`
}

type PluginDeclarationWithoutAdvancedFields struct {
	Version     string                    `json:"version" yaml:"version,omitempty" validate:"required,version"`
	Type        DifyManifestType          `json:"type" yaml:"type,omitempty" validate:"required,eq=plugin"`
	Description I18nObject                `json:"description" yaml:"description" validate:"required"`
	Label       I18nObject                `json:"label" yaml:"label" validate:"required"`
	Author      string                    `json:"author" yaml:"author,omitempty" validate:"omitempty,max=64"`
	Name        string                    `json:"name" yaml:"name,omitempty" validate:"required,max=128"`
	Icon        string                    `json:"icon" yaml:"icon,omitempty" validate:"required,max=128"`
	CreatedAt   time.Time                 `json:"created_at" yaml:"created_at,omitempty" validate:"required"`
	Resource    PluginResourceRequirement `json:"resource" yaml:"resource,omitempty" validate:"required"`
	Plugins     PluginExtensions          `json:"plugins" yaml:"plugins,omitempty" validate:"required"`
	Meta        PluginMeta                `json:"meta" yaml:"meta,omitempty" validate:"required"`
	Tags        []PluginTag               `json:"tags" yaml:"tags,omitempty" validate:"omitempty,dive,plugin_tag,max=128"`
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
		p.Tags = []PluginTag{}
	}

	return nil
}

type PluginDeclaration struct {
	PluginDeclarationWithoutAdvancedFields `yaml:",inline"`
	Verified                               bool                         `json:"verified" yaml:"verified"`
	Endpoint                               *EndpointProviderDeclaration `json:"endpoint,omitempty" yaml:"endpoint,omitempty" validate:"omitempty"`
	Model                                  *ModelProviderDeclaration    `json:"model,omitempty" yaml:"model,omitempty" validate:"omitempty"`
	Tool                                   *ToolProviderDeclaration     `json:"tool,omitempty" yaml:"tool,omitempty" validate:"omitempty"`
}

func (p *PluginDeclaration) UnmarshalJSON(data []byte) error {
	// First unmarshal the embedded struct
	if err := json.Unmarshal(data, &p.PluginDeclarationWithoutAdvancedFields); err != nil {
		return err
	}

	// Then unmarshal the remaining fields
	type PluginExtra struct {
		Verified bool                         `json:"verified"`
		Endpoint *EndpointProviderDeclaration `json:"endpoint,omitempty"`
		Model    *ModelProviderDeclaration    `json:"model,omitempty"`
		Tool     *ToolProviderDeclaration     `json:"tool,omitempty"`
	}

	var extra PluginExtra
	if err := json.Unmarshal(data, &extra); err != nil {
		return err
	}

	p.Verified = extra.Verified
	p.Endpoint = extra.Endpoint
	p.Model = extra.Model
	p.Tool = extra.Tool

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
	PluginNameRegex               = regexp.MustCompile(`^[a-z0-9_-]{1,128}$`)
	AuthorRegex                   = regexp.MustCompile(`^[a-z0-9_-]{1,64}$`)
	PluginDeclarationVersionRegex = regexp.MustCompile(`^\d{1,4}(\.\d{1,4}){1,3}(-\w{1,16})?$`)
)

func isVersion(fl validator.FieldLevel) bool {
	// version format must be like x.x.x, at least 2 digits and most 5 digits, and it can be ends with a letter
	value := fl.Field().String()
	return PluginDeclarationVersionRegex.MatchString(value)
}

func isPluginName(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return PluginNameRegex.MatchString(value)
}

func (p *PluginDeclaration) Identity() string {
	return parser.MarshalPluginID(p.Author, p.Name, p.Version)
}

func (p *PluginDeclaration) ManifestValidate() error {
	if p.Endpoint == nil && p.Model == nil && p.Tool == nil {
		return fmt.Errorf("at least one of endpoint, model, or tool must be provided")
	}

	if p.Model != nil && p.Tool != nil {
		return fmt.Errorf("model and tool cannot be provided at the same time")
	}

	if p.Model != nil && p.Endpoint != nil {
		return fmt.Errorf("model and endpoint cannot be provided at the same time")
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
			deep_copied_description := p.Description
			p.Model.Description = &deep_copied_description
		}
	}

	if p.Tags == nil {
		p.Tags = []PluginTag{}
	}
}

func init() {
	// init validator
	validators.GlobalEntitiesValidator.RegisterValidation("version", isVersion)
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
