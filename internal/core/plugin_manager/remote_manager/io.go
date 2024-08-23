package remote_manager

import (
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/panjf2000/gnet/v2"
)

func (r *RemotePluginRuntime) Listen(session_id string) *entities.Broadcast[plugin_entities.SessionMessage] {
	listener := entities.NewBroadcast[plugin_entities.SessionMessage]()
	listener.OnClose(func() {
		r.removeCallback(session_id)
	})

	r.addCallback(session_id, func(data []byte) {
		// unmarshal the session message
		chunk, err := parser.UnmarshalJsonBytes[plugin_entities.SessionMessage](data)
		if err != nil {
			log.Error("unmarshal json failed: %s, failed to parse session message", err.Error())
			return
		}
		// set the runtime type
		chunk.RuntimeType = r.Type()

		listener.Send(chunk)
	})

	return listener
}

func (r *RemotePluginRuntime) Write(session_id string, data []byte) {
	r.conn.AsyncWrite(append(data, '\n'), func(c gnet.Conn, err error) error {
		return nil
	})
}
