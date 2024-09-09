package positive_manager

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/basic_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
)

type PositivePluginRuntime struct {
	basic_manager.BasicPluginRuntime

	LocalPackagePath string
	WorkingPath      string
	// plugin decoder used to manage the plugin
	Decoder decoder.PluginDecoder

	InnerChecksum string
}
