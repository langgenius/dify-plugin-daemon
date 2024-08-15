package positive_manager

import "github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"

type PositivePluginRuntime struct {
	LocalPackagePath string
	WorkingPath      string
	// plugin decoder used to manage the plugin
	Decoder decoder.PluginDecoder

	checksum string
}
