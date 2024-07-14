package entities

import (
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

type PluginConfiguration struct {
	Version   string                      `json:"version"`
	Author    string                      `json:"author"`
	Name      string                      `json:"name"`
	CreatedAt time.Time                   `json:"createdAt"`
	Module    string                      `json:"module"`
	Resource  PluginConfigurationResource `json:"resource"`
	Meta      PluginConfigurationMeta     `json:"meta"`
	Plugins   []string                    `json:"plugins"`
}

func (p *PluginConfiguration) Identity() string {
	return parser.MarshalPluginIdentity(p.Name, p.Version)
}

type PluginConfigurationResource struct {
	Memory     int64                         `json:"memory"`
	Storage    int64                         `json:"storage"`
	Permission PluginConfigurationPermission `json:"permission"`
}

type PluginConfigurationMeta struct {
	Version string   `json:"version"`
	Arch    []string `json:"arch"`
	Runner  struct {
		Language string `json:"language"`
		Version  string `json:"version"`
	} `json:"runner"`
}

type PluginExtension struct {
	Tool  bool `json:"tool"`
	Model bool `json:"model"`
}

type PluginConfigurationPermission struct {
	Model PluginConfigurationPermissionModel `json:"model"`
	Tool  PluginConfigurationPermissionTool  `json:"tool"`
}

type PluginConfigurationPermissionModel struct {
	Enabled       bool `json:"enabled"`
	LLM           bool `json:"llm"`
	TextEmbedding bool `json:"text_embedding"`
	Rerank        bool `json:"rerank"`
	TTS           bool `json:"tts"`
	STT           bool `json:"stt"`
}

type PluginConfigurationPermissionTool struct {
	Enabled bool `json:"enabled"`
}
