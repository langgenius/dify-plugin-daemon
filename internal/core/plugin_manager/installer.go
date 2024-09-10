package plugin_manager

import "github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"

type Installer interface {
	Install(decoder decoder.PluginDecoder) error
}
