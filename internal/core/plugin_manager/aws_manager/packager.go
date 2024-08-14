package aws_manager

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
)

type Packager struct {
	runtime entities.PluginRuntime
	decoder decoder.PluginDecoder
}

func NewPackager(runtime entities.PluginRuntime, decoder decoder.PluginDecoder) *Packager {
	return &Packager{
		runtime: runtime,
		decoder: decoder,
	}
}

// Pack takes a plugin and packs it into a tar file with dockerfile inside
// for the
func (p *Packager) Pack()
