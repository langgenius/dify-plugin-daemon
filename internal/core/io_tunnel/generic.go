package io_tunnel

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/backwards_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/backwards_invocation/transaction"
	"github.com/langgenius/dify-plugin-daemon/internal/core/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/stream"
)

const (
	writeRetryBudget   = 6 * time.Second
	writeRetryInterval = 300 * time.Millisecond
)

func isRecoverableWriteErr(err error) bool {
	return errors.Is(err, local_runtime.ErrNoProperInstance) || local_runtime.IsInstanceDeadErr(err)
}

func GenericInvokePlugin[Req any, Rsp any](
	session *session_manager.Session,
	request *Req,
	response_buffer_size int,
) (*stream.Stream[Rsp], error) {
	recorder := newPluginInvocationRecorder(session)
	outcome := &pluginInvocationOutcomeTracker{}

	runtime := session.Runtime()
	if runtime == nil {
		recorder.record(pluginInvocationOutcomeError)
		return nil, errors.New("plugin runtime not found")
	}

	response := stream.NewStream[Rsp](response_buffer_size)
	response.OnClose(func() {
		recorder.record(outcome.outcome())
	})

	onChunk := func(chunk plugin_entities.SessionMessage) {
		switch chunk.Type {
		case plugin_entities.SESSION_MESSAGE_TYPE_STREAM:
			chunk, err := parser.UnmarshalJsonBytes[Rsp](chunk.Data)
			if err != nil {
				outcome.markError()
				response.WriteError(errors.New(parser.MarshalJson(map[string]string{
					"error_type": "unmarshal_error",
					"message":    fmt.Sprintf("unmarshal json failed: %s", err.Error()),
				})))
				response.Close()
				return
			} else {
				response.WriteBlocking(chunk)
			}
		case plugin_entities.SESSION_MESSAGE_TYPE_INVOKE:
			if runtime.Type() == plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS {
				outcome.markError()
				response.WriteError(errors.New(parser.MarshalJson(map[string]string{
					"error_type": "serverless_event_not_supported",
					"message":    "serverless event is not supported by full duplex",
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
				outcome.markError()
				response.WriteError(errors.New(parser.MarshalJson(map[string]string{
					"error_type": "invoke_dify_error",
					"message":    fmt.Sprintf("invoke dify failed: %s", err.Error()),
				})))
				response.Close()
				return
			}
		case plugin_entities.SESSION_MESSAGE_TYPE_END:
			outcome.markSuccess()
			response.Close()
		case plugin_entities.SESSION_MESSAGE_TYPE_ERROR:
			outcome.markError()
			e, err := parser.UnmarshalJsonBytes[plugin_entities.ErrorResponse](chunk.Data)
			if err != nil {
				break
			}
			response.WriteError(errors.New(e.Error()))
			response.Close()
		default:
			outcome.markError()
			response.WriteError(errors.New(parser.MarshalJson(map[string]string{
				"error_type": "unknown_stream_message_type",
				"message":    "unknown stream message type: " + string(chunk.Type),
			})))
			response.Close()
		}
	}

	pluginMap := GetInvokePluginMap(
		session,
		request,
	)

	isLocal := runtime.Type() == plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL
	deadline := time.Now().Add(writeRetryBudget)
	for {
		listener, err := runtime.Listen(session.ID)
		if err == nil {
			var closeOnce sync.Once
			closeListener := func() {
				closeOnce.Do(listener.Close)
			}

			listener.Listen(onChunk)

			err = session.Write(
				session_manager.PLUGIN_IN_STREAM_EVENT_REQUEST,
				session.Action,
				pluginMap,
			)
			if err == nil {
				response.OnClose(closeListener)
				return response, nil
			}

			closeListener()
		}

		if !isLocal || !isRecoverableWriteErr(err) || time.Now().After(deadline) {
			recorder.record(pluginInvocationOutcomeError)
			return nil, errors.Join(err, errors.New("failed to write request"))
		}

		time.Sleep(writeRetryInterval)
	}
}
