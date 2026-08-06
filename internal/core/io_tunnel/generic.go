package io_tunnel

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/backwards_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/backwards_invocation/transaction"
	"github.com/langgenius/dify-plugin-daemon/internal/core/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
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

func pluginInvocationLogLabels(session *session_manager.Session) (pluginID string, runtimeType string) {
	pluginID = "unknown"
	runtimeType = "unknown"
	if session == nil || session.Runtime() == nil {
		return pluginID, runtimeType
	}
	runtime := session.Runtime()
	if identity, err := runtime.Identity(); err == nil {
		pluginID = identity.String()
	}
	runtimeType = string(runtime.Type())
	return pluginID, runtimeType
}

func GenericInvokePlugin[Req any, Rsp any](
	session *session_manager.Session,
	request *Req,
	response_buffer_size int,
) (*stream.Stream[Rsp], error) {
	recorder := newPluginInvocationRecorder(session)
	outcome := &pluginInvocationOutcomeTracker{}

	response, err := invokePluginOnce[Req, Rsp](
		session.RequestContext(),
		session,
		request,
		response_buffer_size,
		outcome,
		func() {
			recorder.record(outcome.outcome())
		},
	)
	if err != nil {
		recorder.record(pluginInvocationOutcomeError)
		return nil, err
	}
	return response, nil
}

func invokePluginOnce[Req any, Rsp any](
	ctx context.Context,
	session *session_manager.Session,
	request *Req,
	responseBufferSize int,
	outcome *pluginInvocationOutcomeTracker,
	onClose func(),
) (*stream.Stream[Rsp], error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	runtime := session.Runtime()
	if runtime == nil {
		return nil, errors.New("plugin runtime not found")
	}

	response := stream.NewStream[Rsp](responseBufferSize)
	if onClose != nil {
		response.OnClose(onClose)
	}
	stopCloseOnCancel := context.AfterFunc(ctx, response.Close)
	response.OnClose(func() {
		stopCloseOnCancel()
	})

	onChunk := func(chunk plugin_entities.SessionMessage) {
		switch chunk.Type {
		case plugin_entities.SESSION_MESSAGE_TYPE_STREAM:
			chunk, err := parser.UnmarshalJsonBytes[Rsp](chunk.Data)
			if err != nil {
				outcome.markError()
				response.CloseWithError(errors.New(parser.MarshalJson(map[string]string{
					"error_type": "unmarshal_error",
					"message":    fmt.Sprintf("unmarshal json failed: %s", err.Error()),
				})))
				return
			} else {
				response.WriteBlocking(chunk)
			}
		case plugin_entities.SESSION_MESSAGE_TYPE_INVOKE:
			if runtime.Type() == plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS {
				outcome.markError()
				response.CloseWithError(errors.New(parser.MarshalJson(map[string]string{
					"error_type": "serverless_event_not_supported",
					"message":    "serverless event is not supported by full duplex",
				})))
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
				response.CloseWithError(errors.New(parser.MarshalJson(map[string]string{
					"error_type": "invoke_dify_error",
					"message":    fmt.Sprintf("invoke dify failed: %s", err.Error()),
				})))
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
			pluginID, runtimeType := pluginInvocationLogLabels(session)
			log.ErrorContext(
				session.RequestContext(),
				"plugin returned error response",
				"session_id", session.ID,
				"action", session.Action,
				"plugin", pluginID,
				"runtime_type", runtimeType,
				"error_type", e.ErrorType,
				"message", e.Message,
				"args", e.Args,
			)
			response.CloseWithError(errors.New(e.Error()))
		default:
			outcome.markError()
			response.CloseWithError(errors.New(parser.MarshalJson(map[string]string{
				"error_type": "unknown_stream_message_type",
				"message":    "unknown stream message type: " + string(chunk.Type),
			})))
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
			if err := ctx.Err(); err != nil {
				closeListener()
				stopCloseOnCancel()
				return nil, err
			}

			err = session.WriteContext(
				ctx,
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
			stopCloseOnCancel()
			return nil, errors.Join(err, errors.New("failed to write request"))
		}

		retryTimer := time.NewTimer(writeRetryInterval)
		select {
		case <-ctx.Done():
			retryTimer.Stop()
			stopCloseOnCancel()
			return nil, ctx.Err()
		case <-retryTimer.C:
		}
	}
}
