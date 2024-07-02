package local_manager

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/stdio_holder"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
)

func (r *LocalPluginRuntime) Listen(session_id string) *entities.BytesIOListener {
	listener := entities.NewIOListener[[]byte]()
	listener_id := stdio_holder.OnStdioEvent(r.io_identity, func(b []byte) {
		listener.Write(b)
	})
	listener.OnClose(func() {
		stdio_holder.RemoveStdioListener(r.io_identity, listener_id)
	})
	return listener
}

func (r *LocalPluginRuntime) Write(session_id string, data []byte) {
	stdio_holder.Write(r.io_identity, data)
}

func (r *LocalPluginRuntime) Request(session_id string, data []byte) ([]byte, error) {
	return nil, nil
}
