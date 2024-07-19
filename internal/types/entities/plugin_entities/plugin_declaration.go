package plugin_entities

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

type DifyManifestType string

const (
	PluginType DifyManifestType = "plugin"
)

type PluginPermissionRequirement struct {
	Tool  *PluginPermissionToolRequirement  `json:"tool" yaml:"tool" validate:"omitempty"`
	Model *PluginPermissionModelRequirement `json:"model" yaml:"model" validate:"omitempty"`
	Node  *PluginPermissionNodeRequirement  `json:"node" yaml:"node" validate:"omitempty"`
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
}

type PluginPermissionNodeRequirement struct {
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
	Version   string                     `json:"version" yaml:"version" validate:"required"`
	Type      DifyManifestType           `json:"type" yaml:"type" validate:"required,eq=plugin"`
	Author    string                     `json:"author" yaml:"author" validate:"required"`
	Name      string                     `json:"name" yaml:"name" validate:"required" enum:"plugin"`
	CreatedAt time.Time                  `json:"created_at" yaml:"created_at" validate:"required"`
	Resource  PluginResourceRequirement  `json:"resource" yaml:"resource" validate:"required"`
	Plugins   []string                   `json:"plugins" yaml:"plugins" validate:"required"`
	Execution PluginDeclarationExecution `json:"execution" yaml:"execution" validate:"required"`
}

func (p *PluginDeclaration) Identity() string {
	return parser.MarshalPluginIdentity(p.Name, p.Version)
}

var (
	plugin_declaration_validator = validator.New()
)

func init() {
	// init validator
	plugin_declaration_validator.RegisterValidation("plugin_declaration_platform_arch", isPluginDeclarationPlatformArch)
}

func (p *PluginDeclaration) Validate() error {
	return plugin_declaration_validator.Struct(p)
}

func UnmarshalPluginDeclarationFromYaml(data []byte) (*PluginDeclaration, error) {
	obj, err := parser.UnmarshalYamlBytes[PluginDeclaration](data)
	if err != nil {
		return nil, err
	}

	if err := obj.Validate(); err != nil {
		return nil, err
	}

	return &obj, nil
}

func UnmarshalPluginDeclarationFromJSON(data []byte) (*PluginDeclaration, error) {
	obj, err := parser.UnmarshalJsonBytes[PluginDeclaration](data)
	if err != nil {
		return nil, err
	}

	if err := obj.Validate(); err != nil {
		return nil, err
	}

	return &obj, nil
}
