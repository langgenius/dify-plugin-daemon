package install_service

import (
	controlpanel "github.com/langgenius/dify-plugin-daemon/internal/core/control_panel"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

func HandleLocalInstallationStream(
	pkgFile []byte,
	pluginDecoder decoder.PluginDecoder,
) (*stream.Stream[controlpanel.InstallLocalPluginResponse], error) {
	return nil, nil
}
