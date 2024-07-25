package remote_manager

import (
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/panjf2000/gnet/v2"
)

func (r *RemotePluginRuntime) Listen(session_id string) *entities.BytesIOListener {
	listener := entities.NewIOListener[[]byte]()
	listener.OnClose(func() {
		r.removeCallback(session_id)
	})

	r.addCallback(session_id, func(data []byte) {
		listener.Emit(data)
	})

	return listener
}

func (r *RemotePluginRuntime) Write(session_id string, data []byte) {
	r.conn.AsyncWrite(append(data, '\n'), func(c gnet.Conn, err error) error {
		return nil
	})
}
