package service

import (
	"errors"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	routinepkg "github.com/langgenius/dify-plugin-daemon/pkg/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/metrics"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/stream"
)

// baseSSEService is a helper function to handle SSE service
// it accepts a generator function that returns a stream response to gin context
func baseSSEService[R any](
	generator func() (*stream.Stream[R], error),
	ctx *gin.Context,
	max_timeout_seconds int,
	onCompletion func(status string, duration float64),
) {
	startTime := time.Now()
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
		duration := time.Since(startTime).Seconds()
		if onCompletion != nil {
			onCompletion("error", duration)
		}
		close(done)
		return
	}

	routine.Submit(routinepkg.Labels{
		routinepkg.RoutineLabelKeyModule: "service",
		routinepkg.RoutineLabelKeyMethod: "baseSSEService",
	}, func() {
		status := "success"
		for pluginDaemonResponse.Next() {
			chunk, err := pluginDaemonResponse.Read()
			if err != nil {
				writeData(exception.InvokePluginError(err).ToResponse())
				status = "error"
				break
			}
			writeData(entities.NewSuccessResponse(chunk))
		}

		duration := time.Since(startTime).Seconds()
		if onCompletion != nil {
			onCompletion(status, duration)
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
		duration := time.Since(startTime).Seconds()
		if onCompletion != nil {
			onCompletion("client_disconnect", duration)
		}
		return
	case <-done:
		return
	case <-timer.C:
		writeData(exception.InternalServerError(errors.New("killed by timeout")).ToResponse())
		duration := time.Since(startTime).Seconds()
		if onCompletion != nil {
			onCompletion("timeout", duration)
		}
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
	startTime := time.Now()

	session, err := createSession(
		request,
		access_type,
		access_action,
		ctx.GetString("cluster_id"),
		ctx.Request.Context(),
	)
	if err != nil {
		duration := time.Since(startTime).Seconds()
		recordPluginInvocationMetrics(request, session, access_type, access_action, "error", duration)
		ctx.JSON(500, exception.InternalServerError(err).ToResponse())
		return
	}
	defer session.Close(session_manager.CloseSessionPayload{
		IgnoreCache: false,
	})

	baseSSEService(
		func() (*stream.Stream[R], error) {
			pluginID, runtimeType := getPluginMetricLabels(session)

			metrics.PluginInvocationsActive.WithLabelValues(
				pluginID,
				string(access_type),
				runtimeType,
			).Inc()

			return generator(session)
		},
		ctx,
		max_timeout_seconds,
		func(status string, duration float64) {
			pluginID, runtimeType := getPluginMetricLabels(session)

			metrics.PluginInvocationsTotal.WithLabelValues(
				pluginID,
				string(access_type),
				runtimeType,
				string(access_action),
				status,
			).Inc()
			metrics.PluginInvocationDuration.WithLabelValues(
				pluginID,
				string(access_type),
				runtimeType,
				string(access_action),
			).Observe(duration)

			metrics.PluginInvocationsActive.WithLabelValues(
				pluginID,
				string(access_type),
				runtimeType,
			).Dec()
		},
	)
}

func getPluginMetricLabels(session *session_manager.Session) (pluginID, runtimeType string) {
	pluginID = "unknown"
	runtimeType = "unknown"

	if session != nil && session.Runtime() != nil {
		pluginRuntime := session.Runtime()
		if identity, err := pluginRuntime.Identity(); err == nil {
			pluginID = identity.PluginID()
		}
		runtimeType = string(pluginRuntime.Type())
	}

	return
}

func recordPluginInvocationMetrics[T any](
	request *plugin_entities.InvokePluginRequest[T],
	session *session_manager.Session,
	access_type access_types.PluginAccessType,
	access_action access_types.PluginAccessAction,
	status string,
	duration float64,
) {
	pluginID, runtimeType := getPluginMetricLabels(session)

	metrics.PluginInvocationsTotal.WithLabelValues(
		pluginID,
		string(access_type),
		runtimeType,
		string(access_action),
		status,
	).Inc()

	metrics.PluginInvocationDuration.WithLabelValues(
		pluginID,
		string(access_type),
		runtimeType,
		string(access_action),
	).Observe(duration)
}
