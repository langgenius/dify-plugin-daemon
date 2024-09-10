package installer

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

func LocalInstaller(decoder decoder.PluginDecoder) (*stream.Stream[PluginInstallResponse], error) {
	return nil, nil
}
