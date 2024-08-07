package local_manager

import "github.com/langgenius/dify-plugin-daemon/internal/types/entities"

type LocalPluginRuntime struct {
	entities.PluginRuntime
	CWD string

	io_identity string
	w           chan bool

	checksum string
}
