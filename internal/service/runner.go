package service

import (
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

// baseSSEService is a helper function to handle SSE service
// it accepts a generator function that returns a stream response to gin context
func baseSSEService[R any](
	generator func() (*stream.StreamResponse[R], error),
	ctx *gin.Context,
	max_timeout_seconds int,
) {
	writer := ctx.Writer
	writer.WriteHeader(200)
	writer.Header().Set("Content-Type", "text/event-stream")

	done := make(chan bool)
	done_closed := new(int32)
	closed := new(int32)

	write_data := func(data interface{}) {
		if atomic.LoadInt32(closed) == 1 {
			return
		}
		writer.Write([]byte("data: "))
		writer.Write(parser.MarshalJsonBytes(data))
		writer.Write([]byte("\n\n"))
		writer.Flush()
	}

	plugin_daemon_response, err := generator()

	if err != nil {
		write_data(entities.NewErrorResponse(-500, err.Error()))
		close(done)
		return
	}

	routine.Submit(func() {
		for plugin_daemon_response.Next() {
			chunk, err := plugin_daemon_response.Read()
			if err != nil {
				write_data(entities.NewErrorResponse(-500, err.Error()))
				break
			}
			write_data(entities.NewSuccessResponse(chunk))
		}

		if atomic.CompareAndSwapInt32(done_closed, 0, 1) {
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
		plugin_daemon_response.Close()
		return
	case <-done:
		return
	case <-timer.C:
		write_data(entities.NewErrorResponse(-500, "killed by timeout"))
		if atomic.CompareAndSwapInt32(done_closed, 0, 1) {
			close(done)
		}
		return
	}
}
