package transaction

import (
	"io"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/aws_manager"
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
	runtime *aws_manager.AWSPluginRuntime,
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

	data.RuntimeType = plugin_entities.PLUGIN_RUNTIME_TYPE_AWS
	data.SessionWriter = writer

	// send the data to the plugin runtime
	if err := runtime.PushRequest(session_id, data); err != nil {
		log.Error("push request failed: %s", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	select {
	case <-writer.done:
		return
	case <-time.After(h.max_timeout):
		return
	}
}
