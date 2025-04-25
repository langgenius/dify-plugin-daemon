package service

import (
	"errors"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
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

	routine.Submit(map[string]string{
		"module":   "service",
		"function": "baseSSEService",
	}, func() {
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

func baseSSEWithSession[T any, R any](
	generator func(*session_manager.Session) (*stream.Stream[R], error),
	access_type access_types.PluginAccessType,
	access_action access_types.PluginAccessAction,
	request *plugin_entities.InvokePluginRequest[T],
	ctx *gin.Context,
	max_timeout_seconds int,
) {
	session, err := createSession(
		request,
		access_type,
		access_action,
		ctx.GetString("cluster_id"),
	)
	if err != nil {
		ctx.JSON(500, exception.InternalServerError(err).ToResponse())
		return
	}
	defer session.Close(session_manager.CloseSessionPayload{
		IgnoreCache: false,
	})

	baseSSEService(
		func() (*stream.Stream[R], error) {
			return generator(session)
		},
		ctx,
		max_timeout_seconds,
	)
}
