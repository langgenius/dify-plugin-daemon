package plugin_daemon

import (
	"errors"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

func getInvokeToolMap(
	session *session_manager.Session,
	action PluginAccessAction,
	request *requests.RequestInvokeTool,
) map[string]any {
	req := getBasicPluginAccessMap(session.ID(), session.UserID(), PLUGIN_ACCESS_TYPE_TOOL, action)
	data := req["data"].(map[string]any)

	data["provider"] = request.Provider
	data["tool"] = request.Tool
	data["parameters"] = request.ToolParameters
	data["credentials"] = request.Credentials

	return req
}

func InvokeTool(
	session *session_manager.Session,
	request *requests.RequestInvokeTool,
) (
	*stream.StreamResponse[plugin_entities.ToolResponseChunk], error,
) {
	runtime := plugin_manager.Get(session.PluginIdentity())
	if runtime == nil {
		return nil, errors.New("plugin not found")
	}

	response := stream.NewStreamResponse[plugin_entities.ToolResponseChunk](512)

	listener := runtime.Listen(session.ID())
	listener.AddListener(func(message []byte) {
		chunk, err := parser.UnmarshalJsonBytes[plugin_entities.SessionMessage](message)
		if err != nil {
			log.Error("unmarshal json failed: %s", err.Error())
			return
		}

		switch chunk.Type {
		case plugin_entities.SESSION_MESSAGE_TYPE_STREAM:
			chunk, err := parser.UnmarshalJsonBytes[plugin_entities.ToolResponseChunk](chunk.Data)
			if err != nil {
				log.Error("unmarshal json failed: %s", err.Error())
				return
			}
			response.Write(chunk)
		case plugin_entities.SESSION_MESSAGE_TYPE_INVOKE:
			invokeDify(runtime, session, chunk.Data)
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

	runtime.Write(session.ID(), []byte(parser.MarshalJson(
		getInvokeToolMap(session, PLUGIN_ACCESS_ACTION_INVOKE_TOOL, request)),
	))

	return response, nil
}
