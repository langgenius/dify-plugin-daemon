package plugin_entities

import (
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

type PluginDeclarationWithoutAdvancedFields struct {
	Version   string                    `json:"version" yaml:"version,omitempty" validate:"required,version"`
	Type      DifyManifestType          `json:"type" yaml:"type,omitempty" validate:"required,eq=plugin"`
	Author    string                    `json:"author" yaml:"author,omitempty" validate:"required,max=128"`
	Name      string                    `json:"name" yaml:"name,omitempty" validate:"required,max=128"`
	Icon      string                    `json:"icon" yaml:"icon,omitempty" validate:"required,max=128"`
	Label     I18nObject                `json:"label" yaml:"label" validate:"required"`
	CreatedAt time.Time                 `json:"created_at" yaml:"created_at,omitempty" validate:"required"`
	Resource  PluginResourceRequirement `json:"resource" yaml:"resource,omitempty" validate:"required"`
	Plugins   []string                  `json:"plugins" yaml:"plugins,omitempty" validate:"required,dive,max=128"`
	Meta      PluginMeta                `json:"meta" yaml:"meta,omitempty" validate:"required"`
}

type PluginDeclaration struct {
	PluginDeclarationWithoutAdvancedFields `yaml:",inline"`
	Endpoint                               *EndpointProviderDeclaration `json:"endpoint,omitempty" yaml:"endpoint,omitempty" validate:"omitempty"`
	Model                                  *ModelProviderDeclaration    `json:"model,omitempty" yaml:"model,omitempty" validate:"omitempty"`
	Tool                                   *ToolProviderDeclaration     `json:"tool,omitempty" yaml:"tool,omitempty" validate:"omitempty"`
}

var (
	PluginNameRegex               = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,128}$`)
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
	return parser.MarshalPluginID(p.Name, p.Version)
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

	return &obj, nil
}
