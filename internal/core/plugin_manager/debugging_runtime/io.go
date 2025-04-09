package debugging_runtime

import (
	"encoding/json"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/panjf2000/gnet/v2"
)

func (r *RemotePluginRuntime) Listen(session_id string) *entities.Broadcast[plugin_entities.SessionMessage] {
	listener := entities.NewBroadcast[plugin_entities.SessionMessage]()
	listener.OnClose(func() {
		// execute in new goroutine to avoid deadlock
		routine.Submit(map[string]string{
			"module": "debugging_runtime",
			"method": "removeMessageCallbackHandler",
		}, func() {
			r.removeMessageCallbackHandler(session_id)
			r.removeSessionMessageCloser(session_id)
		})
	})

	// add session message closer to avoid unexpected connection closed
	r.addSessionMessageCloser(session_id, func() {
		listener.Send(plugin_entities.SessionMessage{
			Type: plugin_entities.SESSION_MESSAGE_TYPE_ERROR,
			Data: json.RawMessage(parser.MarshalJson(plugin_entities.ErrorResponse{
				ErrorType: exception.PluginConnectionClosedError,
				Message:   "Connection closed unexpectedly",
				Args:      map[string]any{},
			})),
		})
	})

	r.addMessageCallbackHandler(session_id, func(data []byte) {
		// unmarshal the session message
		chunk, err := parser.UnmarshalJsonBytes[plugin_entities.SessionMessage](data)
		if err != nil {
			log.Error("unmarshal json failed: %s, failed to parse session message", err.Error())
			return
		}

		listener.Send(chunk)
	})

	return listener
}

func (r *RemotePluginRuntime) Write(session_id string, action access_types.PluginAccessAction, data []byte) {
	r.conn.AsyncWrite(append(data, '\n'), func(c gnet.Conn, err error) error {
		return nil
	})
}
