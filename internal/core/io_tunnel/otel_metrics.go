package io_tunnel

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
	gootel "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	pluginInvocationMetricScope = "dify-plugin-daemon/plugin"

	pluginInvocationsMetricName        = "plugin.invocations"
	pluginInvocationDurationMetricName = "plugin.invocation.duration"

	pluginInvocationOutcomeSuccess  = "success"
	pluginInvocationOutcomeError    = "error"
	pluginInvocationOutcomeCanceled = "canceled"
	pluginInvocationUnknownValue    = "unknown"
)

const (
	pluginInvocationStateInFlight int32 = iota
	pluginInvocationStateSuccess
	pluginInvocationStateError
)

type pluginInvocationInstruments struct {
	counter   metric.Int64Counter
	durations metric.Float64Histogram
}

type pluginInvocationRecorder struct {
	session   *session_manager.Session
	startedAt time.Time
	once      sync.Once
}

type pluginInvocationOutcomeTracker struct {
	state int32
}

var (
	pluginInvocationMetricsOnce sync.Once
	pluginInvocationMetrics     *pluginInvocationInstruments
	pluginInvocationMetricsErr  error
)

func newPluginInvocationRecorder(session *session_manager.Session) *pluginInvocationRecorder {
	return &pluginInvocationRecorder{
		session:   session,
		startedAt: time.Now(),
	}
}

func (r *pluginInvocationRecorder) record(outcome string) {
	r.once.Do(func() {
		instruments, err := getPluginInvocationInstruments()
		if err != nil || instruments == nil {
			return
		}

		attrs := buildPluginInvocationAttributes(r.session, outcome)
		ctx := context.Background()
		if r.session != nil {
			ctx = r.session.RequestContext()
		}

		instruments.counter.Add(ctx, 1, metric.WithAttributeSet(attrs))
		instruments.durations.Record(ctx, time.Since(r.startedAt).Seconds(), metric.WithAttributeSet(attrs))
	})
}

func (t *pluginInvocationOutcomeTracker) markSuccess() {
	atomic.CompareAndSwapInt32(&t.state, pluginInvocationStateInFlight, pluginInvocationStateSuccess)
}

func (t *pluginInvocationOutcomeTracker) markError() {
	atomic.StoreInt32(&t.state, pluginInvocationStateError)
}

func (t *pluginInvocationOutcomeTracker) outcome() string {
	switch atomic.LoadInt32(&t.state) {
	case pluginInvocationStateSuccess:
		return pluginInvocationOutcomeSuccess
	case pluginInvocationStateError:
		return pluginInvocationOutcomeError
	default:
		return pluginInvocationOutcomeCanceled
	}
}

func getPluginInvocationInstruments() (*pluginInvocationInstruments, error) {
	pluginInvocationMetricsOnce.Do(func() {
		meter := gootel.Meter(pluginInvocationMetricScope)

		counter, err := meter.Int64Counter(
			pluginInvocationsMetricName,
			metric.WithDescription("Number of plugin runtime invocations handled by the daemon."),
			metric.WithUnit("{call}"),
		)
		if err != nil {
			pluginInvocationMetricsErr = err
			log.Warn("failed to init plugin invocation counter", "error", err)
			return
		}

		durations, err := meter.Float64Histogram(
			pluginInvocationDurationMetricName,
			metric.WithDescription("End-to-end duration of plugin runtime invocations handled by the daemon."),
			metric.WithUnit("s"),
			metric.WithExplicitBucketBoundaries(
				0.005, 0.01, 0.025, 0.05, 0.1,
				0.25, 0.5, 1, 2.5, 5,
				10, 30, 60, 120, 300,
			),
		)
		if err != nil {
			pluginInvocationMetricsErr = err
			log.Warn("failed to init plugin invocation duration histogram", "error", err)
			return
		}

		pluginInvocationMetrics = &pluginInvocationInstruments{
			counter:   counter,
			durations: durations,
		}
	})

	return pluginInvocationMetrics, pluginInvocationMetricsErr
}

func buildPluginInvocationAttributes(session *session_manager.Session, outcome string) attribute.Set {
	pluginID := pluginInvocationUnknownValue
	pluginVersion := pluginInvocationUnknownValue
	runtimeType := pluginInvocationUnknownValue
	accessType := pluginInvocationUnknownValue

	if session != nil {
		if session.PluginUniqueIdentifier != "" {
			if id := session.PluginUniqueIdentifier.PluginID(); id != "" {
				pluginID = id
			}
			if version := string(session.PluginUniqueIdentifier.Version()); version != "" {
				pluginVersion = version
			}
		}
		if session.InvokeFrom != "" {
			accessType = string(session.InvokeFrom)
		}
		if runtime := session.Runtime(); runtime != nil && runtime.Type() != "" {
			runtimeType = string(runtime.Type())
		}
	}

	if outcome == "" {
		outcome = pluginInvocationUnknownValue
	}

	return attribute.NewSet(
		attribute.String("plugin.id", pluginID),
		attribute.String("plugin.version", pluginVersion),
		attribute.String("plugin.runtime_type", runtimeType),
		attribute.String("plugin.access_type", accessType),
		attribute.String("plugin.outcome", outcome),
	)
}

func resetPluginInvocationMetricsForTest() {
	pluginInvocationMetricsOnce = sync.Once{}
	pluginInvocationMetrics = nil
	pluginInvocationMetricsErr = nil
}
