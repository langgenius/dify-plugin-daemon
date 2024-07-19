package service

import (
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

func baseSSEService[T any, R any](
	r *plugin_entities.InvokePluginRequest[T],
	generator func() (*stream.StreamResponse[R], error),
	ctx *gin.Context,
) {
	writer := ctx.Writer
	writer.WriteHeader(200)
	writer.Header().Set("Content-Type", "text/event-stream")

	done := make(chan bool)
	done_closed := new(int32)

	write_data := func(data interface{}) {
		writer.Write([]byte("data: "))
		writer.Write(parser.MarshalJsonBytes(data))
		writer.Write([]byte("\n\n"))
		writer.Flush()
	}

	plugin_daemon_response, err := generator()
	last_response_at := time.Now()

	if err != nil {
		write_data(entities.NewErrorResponse(-500, err.Error()))
		close(done)
		return
	}

	routine.Submit(func() {
		for plugin_daemon_response.Next() {
			last_response_at = time.Now()
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

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-writer.CloseNotify():
			plugin_daemon_response.Close()
			return
		case <-done:
			return
		case <-ticker.C:
			if time.Since(last_response_at) > 30*time.Second {
				write_data(entities.NewErrorResponse(-500, "killed by timeout"))
				if atomic.CompareAndSwapInt32(done_closed, 0, 1) {
					close(done)
				}
				return
			}
		}

	}
}
