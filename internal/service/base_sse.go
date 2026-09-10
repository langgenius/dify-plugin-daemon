package service

import (
	"errors"
	"fmt"
	"sync"
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

const (
	FirstTokenTimeoutErrorType = "FirstTokenTimeoutError"

	firstTokenGraceFloor = 250 * time.Millisecond
	firstTokenGraceRatio = 0.05
)

type firstTokenBudgeter interface {
	FirstTokenBudget() time.Duration
}

type firstTokenCarrier interface {
	CarriesFirstToken() bool
}

type sseOptions struct {
	firstTokenBudget time.Duration
	onFirstToken     func(elapsed time.Duration)
}

func firstTokenGrace(budget time.Duration) time.Duration {
	if grace := time.Duration(float64(budget) * firstTokenGraceRatio); grace > firstTokenGraceFloor {
		return grace
	}
	return firstTokenGraceFloor
}

func firstTokenTimeoutResponse(budget time.Duration) *entities.Response {
	return exception.ErrorWithTypeAndArgs(
		fmt.Sprintf("no token was received within %s", budget),
		FirstTokenTimeoutErrorType,
		map[string]any{
			"first_token_timeout": budget.Seconds(),
			"enforced_by":         "plugin_daemon",
		},
	).ToResponse()
}

func carriesFirstToken(chunk any) bool {
	carrier, ok := chunk.(firstTokenCarrier)
	return !ok || carrier.CarriesFirstToken()
}

// baseSSEService is a helper function to handle SSE service
// it accepts a generator function that returns a stream response to gin context
func baseSSEService[R any](
	generator func() (*stream.Stream[R], error),
	ctx *gin.Context,
	max_timeout_seconds int,
	onCompletion func(status string, duration float64),
	options sseOptions,
) {
	startTime := time.Now()
	writer := ctx.Writer
	writer.WriteHeader(200)
	writer.Header().Set("Content-Type", "text/event-stream")

	done := make(chan bool)
	doneClosed := new(int32)
	closed := new(int32)
	completed := new(int32)

	complete := func(status string) {
		if !atomic.CompareAndSwapInt32(completed, 0, 1) {
			return
		}
		if onCompletion != nil {
			onCompletion(status, time.Since(startTime).Seconds())
		}
	}

	writeLock := sync.Mutex{}
	writeData := func(data interface{}) {
		writeLock.Lock()
		defer writeLock.Unlock()

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
		complete("error")
		close(done)
		return
	}

	firstToken := make(chan struct{})
	firstTokenSeen := new(int32)
	markFirstToken := func() {
		if !atomic.CompareAndSwapInt32(firstTokenSeen, 0, 1) {
			return
		}
		if options.onFirstToken != nil {
			options.onFirstToken(time.Since(startTime))
		}
		close(firstToken)
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
			if carriesFirstToken(chunk) {
				markFirstToken()
			}
		}

		complete(status)

		if atomic.CompareAndSwapInt32(doneClosed, 0, 1) {
			close(done)
		}
	})

	maxTimeout := time.Duration(max_timeout_seconds) * time.Second
	firstTokenTimeout := time.Duration(0)
	if options.firstTokenBudget > 0 {
		firstTokenTimeout = options.firstTokenBudget + firstTokenGrace(options.firstTokenBudget)
	}

	gating := firstTokenTimeout > 0 && firstTokenTimeout < maxTimeout
	timeout := maxTimeout
	if gating {
		timeout = firstTokenTimeout
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	defer func() {
		writeLock.Lock()
		defer writeLock.Unlock()
		atomic.StoreInt32(closed, 1)
	}()

	isDone := func() bool {
		select {
		case <-done:
			return true
		default:
			return false
		}
	}

	stopGating := func() {
		gating = false
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(max(maxTimeout-time.Since(startTime), time.Nanosecond))
	}

	disconnected := writer.CloseNotify()
	firstTokenGate := firstToken
	if !gating {
		firstTokenGate = nil
	}

	for {
		select {
		case <-disconnected:
			pluginDaemonResponse.Close()
			complete("client_disconnect")
			return
		case <-done:
			return
		case <-firstTokenGate:
			firstTokenGate = nil
			stopGating()
		case <-timer.C:
			if isDone() {
				return
			}
			if gating && atomic.LoadInt32(firstTokenSeen) == 1 {
				firstTokenGate = nil
				stopGating()
				continue
			}

			if gating {
				writeData(firstTokenTimeoutResponse(options.firstTokenBudget))
				pluginDaemonResponse.Close()
				complete("first_token_timeout")
			} else {
				writeData(exception.InternalServerError(errors.New("killed by timeout")).ToResponse())
				pluginDaemonResponse.Close()
				complete("timeout")
			}

			if atomic.CompareAndSwapInt32(doneClosed, 0, 1) {
				close(done)
			}
			return
		}
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

	options := sseOptions{}
	if budgeter, ok := any(&request.Data).(firstTokenBudgeter); ok {
		options.firstTokenBudget = budgeter.FirstTokenBudget()
		options.onFirstToken = func(elapsed time.Duration) {
			pluginID, runtimeType := getPluginMetricLabels(session)

			metrics.PluginTimeToFirstToken.WithLabelValues(
				pluginID,
				string(access_type),
				runtimeType,
				string(access_action),
			).Observe(elapsed.Seconds())
		}
	}

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
		options,
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
