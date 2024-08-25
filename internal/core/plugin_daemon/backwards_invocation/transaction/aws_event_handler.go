package transaction

import (
	"io"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/backwards_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

type AWSTransactionHandler struct {
	max_timeout time.Duration
}

func NewAWSTransactionHandler(max_timeout time.Duration) *AWSTransactionHandler {
	return &AWSTransactionHandler{
		max_timeout: max_timeout,
	}
}

type awsTransactionWriteCloser struct {
	gin.ResponseWriter
	done   chan bool
	closed int32
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
		ResponseWriter: ctx.Writer,
		done:           make(chan bool),
	}

	body := ctx.Request.Body
	// read at most 6MB
	bytes, err := io.ReadAll(io.LimitReader(body, 6*1024*1024))
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(err.Error()))
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", "text/event-stream")

	// parse the data
	data, err := parser.UnmarshalJsonBytes[plugin_entities.SessionMessage](bytes)
	if err != nil {
		log.Error("unmarshal json failed: %s, failed to parse session message", err.Error())
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(err.Error()))
		return
	}

	session := session_manager.GetSession(session_id)
	if err != nil {
		log.Error("get session failed: %s", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	aws_response_writer := NewAWSTransactionWriter(session, writer)

	if err := backwards_invocation.InvokeDify(
		session.Declaration,
		session.InvokeFrom,
		session,
		aws_response_writer,
		data.Data,
	); err != nil {
		log.Error("invoke dify failed: %s", err.Error())
	}

	select {
	case <-writer.done:
		return
	case <-time.After(h.max_timeout):
		return
	}
}
