package plugin_manager

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/installer"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

func (p *PluginManager) Install(decoder decoder.PluginDecoder) (*stream.Stream[installer.PluginInstallResponse], error) {
	return p.installer(decoder)
}
