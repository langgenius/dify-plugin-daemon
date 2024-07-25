package local_manager

import (
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
)

func (r *LocalPluginRuntime) Listen(session_id string) *entities.BytesIOListener {
	listener := entities.NewIOListener[[]byte]()
	listener.OnClose(func() {
		RemoveStdioListener(r.io_identity, session_id)
	})
	OnStdioEvent(r.io_identity, session_id, func(b []byte) {
		listener.Emit(b)
	})
	return listener
}

func (r *LocalPluginRuntime) Write(session_id string, data []byte) {
	WriteToStdio(r.io_identity, append(data, '\n'))
}
