package io_tunnel

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
	gootel "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	pluginInvocationMetricScope = "dify-plugin-daemon/plugin"

	pluginInvocationsMetricName        = "plugin.invocations"
	pluginInvocationDurationMetricName = "plugin.invocation.duration"
	serverlessBatchCountMetricName     = "plugin.serverless.batch.count"
	serverlessBatchPayloadMetricName   = "plugin.serverless.batch.payload_bytes"
	serverlessOversizeItemMetricName   = "plugin.serverless.batch.oversize_items"
	serverlessMismatchMetricName       = "plugin.serverless.batch.result_mismatches"

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
	counter                    metric.Int64Counter
	durations                  metric.Float64Histogram
	serverlessBatchCounts      metric.Int64Histogram
	serverlessBatchPayloads    metric.Int64Histogram
	serverlessOversizeItems    metric.Int64Counter
	serverlessResultMismatches metric.Int64Counter
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

func recordServerlessBatchPlan(session *session_manager.Session, payloadBytes []int) {
	instruments, err := getPluginInvocationInstruments()
	if err != nil || instruments == nil {
		return
	}

	ctx := session.RequestContext()
	attrs := buildServerlessBatchAttributes(session)
	instruments.serverlessBatchCounts.Record(
		ctx,
		int64(len(payloadBytes)),
		metric.WithAttributeSet(attrs),
	)
	for _, payloadSize := range payloadBytes {
		instruments.serverlessBatchPayloads.Record(
			ctx,
			int64(payloadSize),
			metric.WithAttributeSet(attrs),
		)
	}
}

func recordServerlessOversizeItem(session *session_manager.Session) {
	instruments, err := getPluginInvocationInstruments()
	if err != nil || instruments == nil {
		return
	}
	instruments.serverlessOversizeItems.Add(
		session.RequestContext(),
		1,
		metric.WithAttributeSet(buildServerlessBatchAttributes(session)),
	)
}

func recordServerlessBatchMismatch(session *session_manager.Session, reason string) {
	instruments, err := getPluginInvocationInstruments()
	if err != nil || instruments == nil {
		return
	}
	attrs := buildServerlessBatchAttributes(session)
	if reason != "" {
		attrs = attribute.NewSet(append(attrs.ToSlice(), attribute.String("batch.mismatch_reason", reason))...)
	}
	instruments.serverlessResultMismatches.Add(
		session.RequestContext(),
		1,
		metric.WithAttributeSet(attrs),
	)
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

		serverlessBatchCounts, err := meter.Int64Histogram(
			serverlessBatchCountMetricName,
			metric.WithDescription("Number of serial Lambda requests planned for one logical plugin invocation."),
			metric.WithUnit("{batch}"),
		)
		if err != nil {
			pluginInvocationMetricsErr = err
			log.Warn("failed to init serverless batch count histogram", "error", err)
			return
		}

		serverlessBatchPayloads, err := meter.Int64Histogram(
			serverlessBatchPayloadMetricName,
			metric.WithDescription("Serialized request payload size for each planned serverless batch."),
			metric.WithUnit("By"),
		)
		if err != nil {
			pluginInvocationMetricsErr = err
			log.Warn("failed to init serverless batch payload histogram", "error", err)
			return
		}

		serverlessOversizeItems, err := meter.Int64Counter(
			serverlessOversizeItemMetricName,
			metric.WithDescription("Number of individual text items that exceed the serverless request limit."),
			metric.WithUnit("{item}"),
		)
		if err != nil {
			pluginInvocationMetricsErr = err
			log.Warn("failed to init serverless oversize item counter", "error", err)
			return
		}

		serverlessResultMismatches, err := meter.Int64Counter(
			serverlessMismatchMetricName,
			metric.WithDescription("Number of invalid or inconsistent serverless batch results."),
			metric.WithUnit("{error}"),
		)
		if err != nil {
			pluginInvocationMetricsErr = err
			log.Warn("failed to init serverless batch mismatch counter", "error", err)
			return
		}

		pluginInvocationMetrics = &pluginInvocationInstruments{
			counter:                    counter,
			durations:                  durations,
			serverlessBatchCounts:      serverlessBatchCounts,
			serverlessBatchPayloads:    serverlessBatchPayloads,
			serverlessOversizeItems:    serverlessOversizeItems,
			serverlessResultMismatches: serverlessResultMismatches,
		}
	})

	return pluginInvocationMetrics, pluginInvocationMetricsErr
}

func buildServerlessBatchAttributes(session *session_manager.Session) attribute.Set {
	pluginID := pluginInvocationUnknownValue
	pluginVersion := pluginInvocationUnknownValue
	accessType := pluginInvocationUnknownValue
	action := pluginInvocationUnknownValue

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
		if session.Action != "" {
			action = string(session.Action)
		}
	}

	return attribute.NewSet(
		attribute.String("plugin.id", pluginID),
		attribute.String("plugin.version", pluginVersion),
		attribute.String("plugin.runtime_type", string(plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS)),
		attribute.String("plugin.access_type", accessType),
		attribute.String("plugin.action", action),
	)
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
