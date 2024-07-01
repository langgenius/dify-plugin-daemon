package entities

import (
	"time"
)

const (
	PLUGIN_RUNTIME_TYPE_LOCAL      = "local"
	PLUGIN_RUNTIME_TYPE_AWS_LAMBDA = "aws_lambda"
)

type PluginRuntime struct {
	Info      PluginRuntimeInfo   `json:"info"`
	State     PluginRuntimeState  `json:"state"`
	Config    PluginConfiguration `json:"config"`
	Connector PluginConnector     `json:"-"`
}

type PluginRuntimeInfo struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Restart bool   `json:"restart"`
}

type PluginRuntimeState struct {
	Restarts     int        `json:"restarts"`
	Active       bool       `json:"active"`
	RelativePath string     `json:"relative_path"`
	ActiveAt     *time.Time `json:"active_at"`
	DeadAt       *time.Time `json:"dead_at"`
	Verified     bool       `json:"verified"`
}

type PluginConfiguration struct {
	Version  string                      `json:"version"`
	Author   string                      `json:"author"`
	Name     string                      `json:"name"`
	Datetime int64                       `json:"datetime"`
	Resource PluginConfigurationResource `json:"resource"`
}

type PluginConfigurationResource struct {
	Memory     int64                         `json:"memory"`
	Storage    int64                         `json:"storage"`
	Permission PluginConfigurationPermission `json:"permission"`
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

type PluginConnector interface {
	OnMessage(func([]byte))
	Read([]byte) int
	Write([]byte) int
}
