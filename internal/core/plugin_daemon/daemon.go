package plugin_daemon

import (
	"errors"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

type ToolResponseChunk = plugin_entities.InvokeToolResponseChunk

func InvokeTool(session *session_manager.Session, provider_name string, tool_name string, tool_parameters map[string]any) (
	*stream.StreamResponse[ToolResponseChunk], error,
) {
	runtime := plugin_manager.Get(session.PluginIdentity())
	if runtime == nil {
		return nil, errors.New("plugin not found")
	}

	response := stream.NewStreamResponse[ToolResponseChunk](512)

	listener := runtime.Listen(session.ID())
	listener.AddListener(func(message []byte) {
		chunk, err := parser.UnmarshalJsonBytes[plugin_entities.StreamMessage](message)
		if err != nil {
			log.Error("unmarshal json failed: %s", err.Error())
			return
		}

		switch chunk.Type {
		case plugin_entities.STREAM_MESSAGE_TYPE_STREAM:
			chunk, err := parser.UnmarshalJsonBytes[ToolResponseChunk](chunk.Data)
			if err != nil {
				log.Error("unmarshal json failed: %s", err.Error())
				return
			}
			response.Write(chunk)
		case plugin_entities.STREAM_MESSAGE_TYPE_END:
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
		map[string]any{
			"provider":   provider_name,
			"tool":       tool_name,
			"parameters": tool_parameters,
			"session_id": session.ID,
		},
	)))

	return response, nil
}
