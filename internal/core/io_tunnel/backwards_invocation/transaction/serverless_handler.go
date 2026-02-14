package transaction

import (
	"io"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/backwards_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
)

type ServerlessTransactionHandler struct {
	maxTimeout time.Duration
}

func NewServerlessTransactionHandler(maxTimeout time.Duration) *ServerlessTransactionHandler {
	return &ServerlessTransactionHandler{
		maxTimeout: maxTimeout,
	}
}

type serverlessTransactionWriteCloser struct {
	done   chan bool
	closed int32

	writer func([]byte) (int, error)
	flush  func()
}

func (a *serverlessTransactionWriteCloser) Write(data []byte) (int, error) {
	return a.writer(data)
}

func (a *serverlessTransactionWriteCloser) Flush() {
	a.flush()
}

func (w *serverlessTransactionWriteCloser) Close() error {
	if atomic.CompareAndSwapInt32(&w.closed, 0, 1) {
		close(w.done)
	}
	return nil
}

func (h *ServerlessTransactionHandler) Handle(ctx *gin.Context, sessionId string) {
	writer := &serverlessTransactionWriteCloser{
		writer: ctx.Writer.Write,
		flush:  ctx.Writer.Flush,
		done:   make(chan bool),
	}

	body := ctx.Request.Body
	// read at most 6MB
	bytes, err := io.ReadAll(io.LimitReader(body, 6*1024*1024))
	if err != nil {
		ctx.Writer.WriteHeader(http.StatusBadRequest)
		ctx.Writer.Write([]byte(err.Error()))
		return
	}

	ctx.Writer.WriteHeader(http.StatusOK)
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")

	plugin_entities.ParsePluginUniversalEvent(
		bytes,
		"",
		func(sessionId string, data []byte) {
			// parse the data
			sessionMessage, err := parser.UnmarshalJsonBytes[plugin_entities.SessionMessage](data)
			if err != nil {
				ctx.Writer.WriteHeader(http.StatusBadRequest)
				ctx.Writer.Write([]byte(err.Error()))
				writer.Close()
				return
			}

			session, err := session_manager.GetSession(sessionId)
			if err != nil {
				ctx.Writer.WriteHeader(http.StatusBadRequest)
				invokePayload, marshalErr := parser.UnmarshalJsonBytes2Map(sessionMessage.Data)
				var backwardsRequestId string
				if marshalErr == nil {
					backwardsRequestId, _ = invokePayload["backwards_request_id"].(string)
				}

				log.ErrorContext(
					ctx.Request.Context(),
					"failed to get session info from cache",
					"session_id", sessionId,
					"backwards_request_id", backwardsRequestId,
					"error", err,
				)

				respData := backwards_invocation.BackwardsInvocationResponseEvent{
					BackwardsRequestId: backwardsRequestId,
					Event:              backwards_invocation.REQUEST_EVENT_ERROR,
					Message:            "failed to get session info from cache",
					Data: map[string]any{
						"error_type": "SessionNotFound",
						"detail":     err.Error(), // 保留“key not found”作为 detail
						"session_id": sessionId,
					},
				}
				_, writeErr := ctx.Writer.Write(parser.MarshalJsonBytes(respData))
				if writeErr != nil {
					log.ErrorContext(
						ctx.Request.Context(),
						"failed to write serverless transaction error response",
						"session_id", sessionId,
						"backwards_request_id", backwardsRequestId,
						"error", writeErr,
					)
				}
				_ = writer.Close()
				return
			}

			// replace trace context, propagate it to gin
			ctxRequestContext := ctx.Request.Context()
			ctxRequestContext = log.WithTrace(ctxRequestContext, session.TraceContext)
			ctxRequestContext = log.WithIdentity(ctxRequestContext, session.IdentityContext)
			ctx.Request = ctx.Request.WithContext(ctxRequestContext)

			// bind the backwards invocation
			plugin_manager := plugin_manager.Manager()
			session.BindBackwardsInvocation(plugin_manager.BackwardsInvocation())

			serverlessResponseWriter := NewServerlessTransactionWriter(session, writer)

			if err := backwards_invocation.InvokeDify(
				session.Declaration,
				session.InvokeFrom,
				session,
				serverlessResponseWriter,
				sessionMessage.Data,
			); err != nil {
				ctx.Writer.WriteHeader(http.StatusInternalServerError)
				ctx.Writer.Write([]byte("failed to parse request"))
				writer.Close()
			}
		},
		func() {},
		func(err string) {
			log.WarnContext(
				ctx.Request.Context(),
				"invoke dify failed, received errors",
				"session_id", sessionId,
				"error", err,
			)
		},
		func(plugin_entities.PluginLogEvent) {}, //log
	)

	select {
	case <-writer.done:
		return
	case <-time.After(h.maxTimeout):
		return
	}
}
