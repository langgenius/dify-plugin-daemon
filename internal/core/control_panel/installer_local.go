package controlpanel

import (
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

func (c *ControlPanel) InstallToLocalFromPkg(
	packageFile []byte,
	pluginDecoder decoder.PluginDecoder,
) (*stream.Stream[InstallLocalPluginResponse], error) {
	// TODO: implement this
	return nil, nil
}
