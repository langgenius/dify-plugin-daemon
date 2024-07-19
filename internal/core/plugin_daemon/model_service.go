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

func getInvokeModelMap(
	session *session_manager.Session,
	action PluginAccessAction,
	request *requests.RequestInvokeLLM,
) map[string]any {
	req := getBasicPluginAccessMap(session.ID(), session.UserID(), PLUGIN_ACCESS_TYPE_MODEL, action)
	data := req["data"].(map[string]any)

	data["provider"] = request.Provider
	data["model"] = request.Model
	data["model_type"] = request.ModelType
	data["model_parameters"] = request.ModelParameters
	data["prompt_messages"] = request.PromptMessages
	data["tools"] = request.Tools
	data["stop"] = request.Stop
	data["stream"] = request.Stream
	data["credentials"] = request.Credentials

	return req
}

func InvokeLLM(
	session *session_manager.Session,
	request *requests.RequestInvokeLLM,
) (
	*stream.StreamResponse[plugin_entities.InvokeModelResponseChunk], error,
) {
	runtime := plugin_manager.Get(session.PluginIdentity())
	if runtime == nil {
		return nil, errors.New("plugin not found")
	}

	response := stream.NewStreamResponse[plugin_entities.InvokeModelResponseChunk](512)

	listener := runtime.Listen(session.ID())
	listener.AddListener(func(message []byte) {
		chunk, err := parser.UnmarshalJsonBytes[plugin_entities.SessionMessage](message)
		if err != nil {
			log.Error("unmarshal json failed: %s", err.Error())
			return
		}

		switch chunk.Type {
		case plugin_entities.SESSION_MESSAGE_TYPE_STREAM:
			chunk, err := parser.UnmarshalJsonBytes[plugin_entities.InvokeModelResponseChunk](chunk.Data)
			if err != nil {
				log.Error("unmarshal json failed: %s", err.Error())
				return
			}
			response.Write(chunk)
		case plugin_entities.SESSION_MESSAGE_TYPE_INVOKE:
			invokeDify(runtime, session, chunk.Data)
		case plugin_entities.SESSION_MESSAGE_TYPE_END:
			response.Close()
		default:
			log.Error("unknown stream message type: %s", chunk.Type)
			response.Close()
		}
	})

	response.OnClose(func() {
		listener.Close()
	})

	runtime.Write(session.ID(), []byte(parser.MarshalJson(
		getInvokeModelMap(
			session,
			PLUGIN_ACCESS_ACTION_INVOKE_LLM,
			request,
		),
	)))

	return response, nil
}
