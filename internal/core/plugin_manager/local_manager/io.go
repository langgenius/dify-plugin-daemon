package local_manager

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/stdio_holder"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
)

func (r *LocalPluginRuntime) Listen(session_id string) *entities.BytesIOListener {
	listener := entities.NewIOListener[[]byte]()
	listener.OnClose(func() {
		stdio_holder.RemoveListener(r.io_identity, session_id)
	})
	stdio_holder.OnEvent(r.io_identity, session_id, func(b []byte) {
		listener.Emit(b)
	})

	return listener
}

func (r *LocalPluginRuntime) Write(session_id string, data []byte) {
	stdio_holder.Write(r.io_identity, append(data, '\n'))
}

func (r *LocalPluginRuntime) Request(session_id string, data []byte) ([]byte, error) {
	return nil, nil
}
