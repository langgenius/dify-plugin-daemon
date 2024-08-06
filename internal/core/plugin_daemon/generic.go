package plugin_daemon

import (
	"errors"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/backwards_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

func genericInvokePlugin[Req any, Rsp any](
	session *session_manager.Session,
	request *Req,
	response_buffer_size int,
	typ access_types.PluginAccessType,
	action access_types.PluginAccessAction,
) (*stream.StreamResponse[Rsp], error) {
	runtime := plugin_manager.GetGlobalPluginManager().Get(session.PluginIdentity())
	if runtime == nil {
		return nil, errors.New("plugin not found")
	}

	response := stream.NewStreamResponse[Rsp](response_buffer_size)

	listener := runtime.Listen(session.ID())
	listener.AddListener(func(message []byte) {
		chunk, err := parser.UnmarshalJsonBytes[plugin_entities.SessionMessage](message)
		if err != nil {
			log.Error("unmarshal json failed: %s", err.Error())
			return
		}

		switch chunk.Type {
		case plugin_entities.SESSION_MESSAGE_TYPE_STREAM:
			chunk, err := parser.UnmarshalJsonBytes[Rsp](chunk.Data)
			if err != nil {
				log.Error("unmarshal json failed: %s", err.Error())
				return
			}
			response.Write(chunk)
		case plugin_entities.SESSION_MESSAGE_TYPE_INVOKE:
			if err := backwards_invocation.InvokeDify(runtime, typ, session, chunk.Data); err != nil {
				log.Error("invoke dify failed: %s", err.Error())
				return
			}
		case plugin_entities.SESSION_MESSAGE_TYPE_END:
			response.Close()
		case plugin_entities.SESSION_MESSAGE_TYPE_ERROR:
			e, err := parser.UnmarshalJsonBytes[plugin_entities.ErrorResponse](chunk.Data)
			if err != nil {
				break
			}
			response.WriteError(errors.New(e.Error))
			response.Close()
		default:
			response.WriteError(errors.New("unknown stream message type: " + string(chunk.Type)))
			response.Close()
		}
	})

	response.OnClose(func() {
		listener.Close()
	})

	session.Write(
		session_manager.PLUGIN_IN_STREAM_EVENT_REQUEST,
		getInvokePluginMap(
			session,
			typ,
			action,
			request,
		),
	)

	return response, nil
}

func getInvokePluginMap(
	session *session_manager.Session,
	typ access_types.PluginAccessType,
	action access_types.PluginAccessAction,
	request any,
) map[string]any {
	req := getBasicPluginAccessMap(session.UserID(), typ, action)
	for k, v := range parser.StructToMap(request) {
		req[k] = v
	}
	return req
}
