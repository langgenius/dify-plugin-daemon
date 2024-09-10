package installer

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

type PluginInstallResponse struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

type Installer func(decoder decoder.PluginDecoder) (*stream.Stream[PluginInstallResponse], error)
