package install_service

import (
	serverless "github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/serverless_connector"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

func HandleServerlessInstallationStream(
	pkgFile []byte,
	pluginDecoder decoder.PluginDecoder,
	stream *stream.Stream[serverless.LaunchFunctionResponse],
) (*stream.Stream[PluginInstallResponse], error) {

}
