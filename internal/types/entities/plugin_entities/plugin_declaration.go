package plugin_entities

import (
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/internal/types/validators"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

type DifyManifestType string

const (
	PluginType DifyManifestType = "plugin"
)

type PluginPermissionRequirement struct {
	Tool    *PluginPermissionToolRequirement    `json:"tool" yaml:"tool" validate:"omitempty"`
	Model   *PluginPermissionModelRequirement   `json:"model" yaml:"model" validate:"omitempty"`
	Node    *PluginPermissionNodeRequirement    `json:"node" yaml:"node" validate:"omitempty"`
	Webhook *PluginPermissionWebhookRequirement `json:"webhook" yaml:"webhook" validate:"omitempty"`
}

func (p *PluginPermissionRequirement) AllowInvokeTool() bool {
	return p != nil && p.Tool != nil && p.Tool.Enabled
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

func (p *PluginPermissionRequirement) AllowRegistryWebhook() bool {
	return p != nil && p.Webhook != nil && p.Webhook.Enabled
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

type PluginPermissionWebhookRequirement struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}

type PluginResourceRequirement struct {
	// Memory in bytes
	Memory int64 `json:"memory" yaml:"memory" validate:"required"`
	// Storage in bytes
	Storage int64 `json:"storage" yaml:"storage" validate:"required"`
	// Permission requirements
	Permission *PluginPermissionRequirement `json:"permission" yaml:"permission" validate:"omitempty"`
}

type PluginDeclarationPlatformArch string

const (
	PLUGIN_PLATFORM_ARCH_AMD64 PluginDeclarationPlatformArch = "amd64"
	PLUGIN_PLATFORM_ARCH_ARM64 PluginDeclarationPlatformArch = "arm64"
)

func isPluginDeclarationPlatformArch(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	switch value {
	case string(PLUGIN_PLATFORM_ARCH_AMD64),
		string(PLUGIN_PLATFORM_ARCH_ARM64):
		return true
	}
	return false
}

type PluginDeclarationMeta struct {
	Version string   `json:"version" yaml:"version" validate:"required"`
	Arch    []string `json:"arch" yaml:"arch" validate:"required,dive,plugin_declaration_platform_arch"`
}

type PluginDeclarationExecution struct {
	Install string `json:"install" yaml:"install" validate:"omitempty"`
	Launch  string `json:"launch" yaml:"launch" validate:"omitempty"`
}

type PluginDeclaration struct {
	Version   string                     `json:"version" yaml:"version" validate:"required,version"`
	Type      DifyManifestType           `json:"type" yaml:"type" validate:"required,eq=plugin"`
	Author    string                     `json:"author" yaml:"author" validate:"required"`
	Name      string                     `json:"name" yaml:"name" validate:"required" enum:"plugin"`
	CreatedAt time.Time                  `json:"created_at" yaml:"created_at" validate:"required"`
	Resource  PluginResourceRequirement  `json:"resource" yaml:"resource" validate:"required"`
	Plugins   []string                   `json:"plugins" yaml:"plugins" validate:"required"`
	Execution PluginDeclarationExecution `json:"execution" yaml:"execution" validate:"required"`
}

var (
	plugin_declaration_version_regex = regexp.MustCompile(`^\d{1,4}(\.\d{1,4}){1,3}(-\w{1,16})?$`)
)

func isVersion(fl validator.FieldLevel) bool {
	// version format must be like x.x.x, at least 2 digits and most 5 digits, can be ends with a letter
	value := fl.Field().String()
	return plugin_declaration_version_regex.MatchString(value)
}

func (p *PluginDeclaration) Identity() string {
	return parser.MarshalPluginIdentity(p.Name, p.Version)
}

func init() {
	// init validator
	validators.GlobalEntitiesValidator.RegisterValidation("plugin_declaration_platform_arch", isPluginDeclarationPlatformArch)
	validators.GlobalEntitiesValidator.RegisterValidation("version", isVersion)
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
