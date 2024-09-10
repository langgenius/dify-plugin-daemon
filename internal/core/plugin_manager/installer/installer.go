package installer

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

type Installer func(decoder decoder.PluginDecoder) (*stream.Stream[string], error)
