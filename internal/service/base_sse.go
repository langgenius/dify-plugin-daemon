package service

import (
	"errors"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

// baseSSEService is a helper function to handle SSE service
// it accepts a generator function that returns a stream response to gin context
func baseSSEService[R any](
	generator func() (*stream.Stream[R], error),
	ctx *gin.Context,
	max_timeout_seconds int,
) {
	writer := ctx.Writer
	writer.WriteHeader(200)
	writer.Header().Set("Content-Type", "text/event-stream")

	done := make(chan bool)
	doneClosed := new(int32)
	closed := new(int32)

	writeData := func(data interface{}) {
		if atomic.LoadInt32(closed) == 1 {
			return
		}
		writer.Write([]byte("data: "))
		writer.Write(parser.MarshalJsonBytes(data))
		writer.Write([]byte("\n\n"))
		writer.Flush()
	}

	pluginDaemonResponse, err := generator()

	if err != nil {
		writeData(exception.InternalServerError(err).ToResponse())
		close(done)
		return
	}

	routine.Submit(func() {
		for pluginDaemonResponse.Next() {
			chunk, err := pluginDaemonResponse.Read()
			if err != nil {
				writeData(exception.InvokePluginError(err).ToResponse())
				break
			}
			writeData(entities.NewSuccessResponse(chunk))
		}

		if atomic.CompareAndSwapInt32(doneClosed, 0, 1) {
			close(done)
		}
	})

	timer := time.NewTimer(time.Duration(max_timeout_seconds) * time.Second)
	defer timer.Stop()

	defer func() {
		atomic.StoreInt32(closed, 1)
	}()

	select {
	case <-writer.CloseNotify():
		pluginDaemonResponse.Close()
		return
	case <-done:
		return
	case <-timer.C:
		writeData(exception.InternalServerError(errors.New("killed by timeout")).ToResponse())
		if atomic.CompareAndSwapInt32(doneClosed, 0, 1) {
			close(done)
		}
		return
	}
}
