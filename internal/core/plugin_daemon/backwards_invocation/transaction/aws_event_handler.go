package transaction

import (
	"io"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/backwards_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

type AWSTransactionHandler struct {
	maxTimeout time.Duration
}

func NewAWSTransactionHandler(maxTimeout time.Duration) *AWSTransactionHandler {
	return &AWSTransactionHandler{
		maxTimeout: maxTimeout,
	}
}

type awsTransactionWriteCloser struct {
	done   chan bool
	closed int32

	writer func([]byte) (int, error)
	flush  func()
}

func (a *awsTransactionWriteCloser) Write(data []byte) (int, error) {
	return a.writer(data)
}

func (a *awsTransactionWriteCloser) Flush() {
	a.flush()
}

func (w *awsTransactionWriteCloser) Close() error {
	if atomic.CompareAndSwapInt32(&w.closed, 0, 1) {
		close(w.done)
	}
	return nil
}

func (h *AWSTransactionHandler) Handle(
	ctx *gin.Context,
	session_id string,
) {
	writer := &awsTransactionWriteCloser{
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
		func(session_id string, data []byte) {
			// parse the data
			sessionMessage, err := parser.UnmarshalJsonBytes[plugin_entities.SessionMessage](data)
			if err != nil {
				ctx.Writer.WriteHeader(http.StatusBadRequest)
				ctx.Writer.Write([]byte(err.Error()))
				writer.Close()
				return
			}

			session := session_manager.GetSession(session_manager.GetSessionPayload{
				ID: session_id,
			})
			if session == nil {
				log.Error("session not found: %s", session_id)
				ctx.Writer.WriteHeader(http.StatusInternalServerError)
				ctx.Writer.Write([]byte("session not found"))
				writer.Close()
				return
			}

			// bind the backwards invocation
			plugin_manager := plugin_manager.Manager()
			session.BindBackwardsInvocation(plugin_manager.BackwardsInvocation())

			awsResponseWriter := NewAWSTransactionWriter(session, writer)

			if err := backwards_invocation.InvokeDify(
				session.Declaration,
				session.InvokeFrom,
				session,
				awsResponseWriter,
				sessionMessage.Data,
			); err != nil {
				ctx.Writer.WriteHeader(http.StatusInternalServerError)
				ctx.Writer.Write([]byte("failed to parse request"))
				writer.Close()
			}
		},
		func() {},
		func(err string) {
			log.Warn("invoke dify failed, received errors: %s", err)
		},
		func(message string) {}, //log
	)

	select {
	case <-writer.done:
		return
	case <-time.After(h.maxTimeout):
		return
	}
}
