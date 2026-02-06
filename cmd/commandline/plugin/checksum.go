package plugin

import (
	"os"

	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

func CalculateChecksum(pluginPath string) {
	var pluginDecoder decoder.PluginDecoder
	if stat, err := os.Stat(pluginPath); err == nil {
		if stat.IsDir() {
			pluginDecoder, err = decoder.NewFSPluginDecoder(pluginPath)
			if err != nil {
				log.Error("failed to create plugin decoder", "plugin_path", pluginPath, "error", err)
				return
			}
		} else {
			bytes, err := os.ReadFile(pluginPath)
			if err != nil {
				log.Error("failed to read plugin file", "plugin_path", pluginPath, "error", err)
				return
			}

			pluginDecoder, err = decoder.NewZipPluginDecoder(bytes)
			if err != nil {
				log.Error("failed to create plugin decoder", "plugin_path", pluginPath, "error", err)
				return
			}
		}
	} else {
		log.Error("failed to get plugin file info", "plugin_path", pluginPath, "error", err)
		return
	}

	checksum, err := pluginDecoder.Checksum()
	if err != nil {
		log.Error("failed to calculate checksum", "plugin_path", pluginPath, "error", err)
		return
	}

	log.Info("plugin checksum", "checksum", checksum)
}
