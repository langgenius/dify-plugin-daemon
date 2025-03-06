package plugin_daemon

import (
	"errors"
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/backwards_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/backwards_invocation/transaction"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func GenericInvokePlugin[Req any, Rsp any](
	session *session_manager.Session,
	request *Req,
	response_buffer_size int,
) (*stream.Stream[Rsp], error) {
	runtime := session.Runtime()
	if runtime == nil {
		return nil, errors.New("plugin runtime not found")
	}

	response := stream.NewStream[Rsp](response_buffer_size)

	listener := runtime.Listen(session.ID)
	listener.Listen(func(chunk plugin_entities.SessionMessage) {
		switch chunk.Type {
		case plugin_entities.SESSION_MESSAGE_TYPE_STREAM:
			chunk, err := parser.UnmarshalJsonBytes[Rsp](chunk.Data)
			if err != nil {
				response.WriteError(errors.New(parser.MarshalJson(map[string]string{
					"error_type": "unmarshal_error",
					"message":    fmt.Sprintf("unmarshal json failed: %s", err.Error()),
				})))
				response.Close()
				return
			} else {
				response.Write(chunk)
			}
		case plugin_entities.SESSION_MESSAGE_TYPE_INVOKE:
			// check if the request contains a aws_event_id
			if runtime.Type() == plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS {
				response.WriteError(errors.New(parser.MarshalJson(map[string]string{
					"error_type": "aws_event_not_supported",
					"message":    "aws event is not supported by full duplex",
				})))
				response.Close()
				return
			}
			if err := backwards_invocation.InvokeDify(
				runtime.Configuration(),
				session.InvokeFrom,
				session,
				transaction.NewFullDuplexEventWriter(session),
				chunk.Data,
			); err != nil {
				response.WriteError(errors.New(parser.MarshalJson(map[string]string{
					"error_type": "invoke_dify_error",
					"message":    fmt.Sprintf("invoke dify failed: %s", err.Error()),
				})))
				response.Close()
				return
			}
		case plugin_entities.SESSION_MESSAGE_TYPE_END:
			response.Close()
		case plugin_entities.SESSION_MESSAGE_TYPE_ERROR:
			e, err := parser.UnmarshalJsonBytes[plugin_entities.ErrorResponse](chunk.Data)
			if err != nil {
				break
			}
			response.WriteError(errors.New(e.Error()))
			response.Close()
		default:
			response.WriteError(errors.New(parser.MarshalJson(map[string]string{
				"error_type": "unknown_stream_message_type",
				"message":    "unknown stream message type: " + string(chunk.Type),
			})))
			response.Close()
		}
	})

	// close the listener if stream outside is closed due to close of connection
	response.OnClose(func() {
		listener.Close()
	})

	session.Write(
		session_manager.PLUGIN_IN_STREAM_EVENT_REQUEST,
		session.Action,
		getInvokePluginMap(
			session,
			request,
		),
	)

	return response, nil
}

func getInvokePluginMap(
	session *session_manager.Session,
	request any,
) map[string]any {
	req := getBasicPluginAccessMap(
		session.UserID,
		session.InvokeFrom,
		session.Action,
	)
	for k, v := range parser.StructToMap(request) {
		req[k] = v
	}
	return req
}
