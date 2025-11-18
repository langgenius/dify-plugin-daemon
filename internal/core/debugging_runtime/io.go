package debugging_runtime

import (
	"encoding/json"
	"errors"

	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	routinepkg "github.com/langgenius/dify-plugin-daemon/pkg/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
	"github.com/panjf2000/gnet/v2"
)

func (r *RemotePluginRuntime) Listen(sessionId string) (*entities.Broadcast[plugin_entities.SessionMessage], error) {
	listener := entities.NewCallbackHandler[plugin_entities.SessionMessage]()
	listener.OnClose(func() {
		// execute in new goroutine to avoid deadlock
		routine.Submit(routinepkg.Labels{
			routinepkg.RoutineLabelKeyModule: "debugging_runtime",
			routinepkg.RoutineLabelKeyMethod: "removeMessageCallbackHandler",
		}, func() {
			r.removeMessageCallbackHandler(sessionId)
			r.removeSessionMessageCloser(sessionId)
		})
	})

	// add session message closer to avoid unexpected connection closed
	r.addSessionMessageCloser(sessionId, func() {
		listener.Send(plugin_entities.SessionMessage{
			Type: plugin_entities.SESSION_MESSAGE_TYPE_ERROR,
			Data: json.RawMessage(parser.MarshalJson(plugin_entities.ErrorResponse{
				ErrorType: exception.PluginConnectionClosedError,
				Message:   "Connection closed unexpectedly",
				Args:      map[string]any{},
			})),
		})
	})

	r.addMessageCallbackHandler(sessionId, func(data []byte) {
		// unmarshal the session message
		chunk, err := parser.UnmarshalJsonBytes[plugin_entities.SessionMessage](data)
		if err != nil {
			log.Error("unmarshal json failed: %s, failed to parse session message", err.Error())
			return
		}

		listener.Send(chunk)
	})

	return listener, nil
}

func (r *RemotePluginRuntime) Write(
	sessionId string,
	action access_types.PluginAccessAction,
	data []byte,
) error {
	if r.conn == nil {
		return errors.New("connection not established")
	}
	return r.conn.AsyncWrite(append(data, '\n'), func(c gnet.Conn, err error) error {
		return err
	})
}
