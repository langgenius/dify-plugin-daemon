package io_tunnel

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/stretchr/testify/require"
	gootel "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

var testPluginUniqueIdentifier = plugin_entities.PluginUniqueIdentifier(
	"author/demo-plugin:1.2.3@0123456789abcdef0123456789abcdef",
)

type fakePluginRuntime struct {
	listener         *entities.Broadcast[plugin_entities.SessionMessage]
	uniqueIdentifier plugin_entities.PluginUniqueIdentifier
	runtimeType      plugin_entities.PluginRuntimeType
	writeFn          func(string, access_types.PluginAccessAction, []byte) error
}

func (f *fakePluginRuntime) Type() plugin_entities.PluginRuntimeType {
	return f.runtimeType
}

func (f *fakePluginRuntime) Configuration() *plugin_entities.PluginDeclaration {
	return &plugin_entities.PluginDeclaration{}
}

func (f *fakePluginRuntime) Identity() (plugin_entities.PluginUniqueIdentifier, error) {
	return f.uniqueIdentifier, nil
}

func (f *fakePluginRuntime) HashedIdentity() (string, error) {
	return plugin_entities.HashedIdentity(f.uniqueIdentifier.String()), nil
}

func (f *fakePluginRuntime) Checksum() (string, error) {
	return f.uniqueIdentifier.Checksum(), nil
}

func (f *fakePluginRuntime) Listen(string) (*entities.Broadcast[plugin_entities.SessionMessage], error) {
	return f.listener, nil
}

func (f *fakePluginRuntime) Write(sessionID string, action access_types.PluginAccessAction, data []byte) error {
	if f.writeFn != nil {
		return f.writeFn(sessionID, action, data)
	}
	return nil
}

func TestGenericInvokePluginRecordsSuccessMetrics(t *testing.T) {
	reader := setupPluginInvocationMetricTest(t)
	listener := entities.NewCallbackHandler[plugin_entities.SessionMessage]()

	runtime := &fakePluginRuntime{
		listener:         listener,
		uniqueIdentifier: testPluginUniqueIdentifier,
		runtimeType:      plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL,
	}
	runtime.writeFn = func(string, access_types.PluginAccessAction, []byte) error {
		go func() {
			time.Sleep(10 * time.Millisecond)
			listener.Send(plugin_entities.SessionMessage{
				Type: plugin_entities.SESSION_MESSAGE_TYPE_END,
			})
		}()
		return nil
	}

	session := newTestSession(t, runtime, access_types.PLUGIN_ACCESS_TYPE_TOOL, access_types.PLUGIN_ACCESS_ACTION_INVOKE_TOOL)
	request := map[string]any{"hello": "world"}

	response, err := GenericInvokePlugin[map[string]any, map[string]any](session, &request, 2)
	require.NoError(t, err)

	for response.Next() {
		_, err := response.Read()
		require.NoError(t, err)
	}

	metrics := collectPluginInvocationMetrics(t, reader)
	expectedAttrs := expectedPluginMetricAttributes(
		pluginInvocationOutcomeSuccess,
		string(plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL),
		string(access_types.PLUGIN_ACCESS_TYPE_TOOL),
	)

	require.Eventually(t, func() bool {
		metrics = collectPluginInvocationMetrics(t, reader)
		counter, ok := pluginInvocationCounterValue(metrics, expectedAttrs)
		if !ok || counter != 1 {
			return false
		}

		histogramCount, ok := pluginInvocationHistogramCount(metrics, expectedAttrs)
		return ok && histogramCount == 1
	}, time.Second, 10*time.Millisecond)
}

func TestGenericInvokePluginRecordsErrorMetrics(t *testing.T) {
	reader := setupPluginInvocationMetricTest(t)

	runtime := &fakePluginRuntime{
		listener:         entities.NewCallbackHandler[plugin_entities.SessionMessage](),
		uniqueIdentifier: testPluginUniqueIdentifier,
		runtimeType:      plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS,
		writeFn: func(string, access_types.PluginAccessAction, []byte) error {
			return errors.New("write failed")
		},
	}

	session := newTestSession(t, runtime, access_types.PLUGIN_ACCESS_TYPE_ENDPOINT, access_types.PLUGIN_ACCESS_ACTION_INVOKE_ENDPOINT)
	request := map[string]any{"hello": "world"}

	response, err := GenericInvokePlugin[map[string]any, map[string]any](session, &request, 2)
	require.Nil(t, response)
	require.Error(t, err)

	metrics := collectPluginInvocationMetrics(t, reader)
	expectedAttrs := expectedPluginMetricAttributes(
		pluginInvocationOutcomeError,
		string(plugin_entities.PLUGIN_RUNTIME_TYPE_SERVERLESS),
		string(access_types.PLUGIN_ACCESS_TYPE_ENDPOINT),
	)

	require.Eventually(t, func() bool {
		metrics = collectPluginInvocationMetrics(t, reader)
		counter, ok := pluginInvocationCounterValue(metrics, expectedAttrs)
		if !ok || counter != 1 {
			return false
		}

		histogramCount, ok := pluginInvocationHistogramCount(metrics, expectedAttrs)
		return ok && histogramCount == 1
	}, time.Second, 10*time.Millisecond)
}

func TestGenericInvokePluginRecordsCanceledMetrics(t *testing.T) {
	reader := setupPluginInvocationMetricTest(t)

	runtime := &fakePluginRuntime{
		listener:         entities.NewCallbackHandler[plugin_entities.SessionMessage](),
		uniqueIdentifier: testPluginUniqueIdentifier,
		runtimeType:      plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL,
	}

	session := newTestSession(t, runtime, access_types.PLUGIN_ACCESS_TYPE_TRIGGER, access_types.PLUGIN_ACCESS_ACTION_SUBSCRIBE_TRIGGER)
	request := map[string]any{"hello": "world"}

	response, err := GenericInvokePlugin[map[string]any, map[string]any](session, &request, 2)
	require.NoError(t, err)

	response.Close()

	metrics := collectPluginInvocationMetrics(t, reader)
	expectedAttrs := expectedPluginMetricAttributes(
		pluginInvocationOutcomeCanceled,
		string(plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL),
		string(access_types.PLUGIN_ACCESS_TYPE_TRIGGER),
	)

	require.Eventually(t, func() bool {
		metrics = collectPluginInvocationMetrics(t, reader)
		counter, ok := pluginInvocationCounterValue(metrics, expectedAttrs)
		if !ok || counter != 1 {
			return false
		}

		histogramCount, ok := pluginInvocationHistogramCount(metrics, expectedAttrs)
		return ok && histogramCount == 1
	}, time.Second, 10*time.Millisecond)
}

func setupPluginInvocationMetricTest(t *testing.T) *sdkmetric.ManualReader {
	t.Helper()

	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	previous := gootel.GetMeterProvider()

	resetPluginInvocationMetricsForTest()
	gootel.SetMeterProvider(provider)
	instruments, err := getPluginInvocationInstruments()
	require.NoError(t, err)
	require.NotNil(t, instruments)

	t.Cleanup(func() {
		resetPluginInvocationMetricsForTest()
		gootel.SetMeterProvider(previous)
		_ = provider.Shutdown(context.Background())
	})

	return reader
}

func newTestSession(
	t *testing.T,
	runtime plugin_entities.PluginRuntimeSessionIOInterface,
	accessType access_types.PluginAccessType,
	action access_types.PluginAccessAction,
) *session_manager.Session {
	t.Helper()

	session := session_manager.NewSession(session_manager.NewSessionPayload{
		TenantID:               "tenant-1",
		UserID:                 "user-1",
		PluginUniqueIdentifier: testPluginUniqueIdentifier,
		InvokeFrom:             accessType,
		Action:                 action,
		RequestContext:         context.Background(),
		IgnoreCache:            true,
	})
	session.BindRuntime(runtime)

	t.Cleanup(func() {
		session.Close(session_manager.CloseSessionPayload{
			IgnoreCache: true,
		})
	})

	return session
}

func collectPluginInvocationMetrics(t *testing.T, reader *sdkmetric.ManualReader) metricdata.ResourceMetrics {
	t.Helper()

	var metrics metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &metrics))
	return metrics
}

func expectedPluginMetricAttributes(outcome, runtimeType, accessType string) map[string]string {
	return map[string]string{
		"plugin.id":           testPluginUniqueIdentifier.PluginID(),
		"plugin.version":      string(testPluginUniqueIdentifier.Version()),
		"plugin.runtime_type": runtimeType,
		"plugin.access_type":  accessType,
		"plugin.outcome":      outcome,
	}
}

func pluginInvocationCounterValue(
	metrics metricdata.ResourceMetrics,
	expectedAttrs map[string]string,
) (int64, bool) {
	for _, scopeMetrics := range metrics.ScopeMetrics {
		for _, metric := range scopeMetrics.Metrics {
			if metric.Name != pluginInvocationsMetricName {
				continue
			}

			sum, ok := metric.Data.(metricdata.Sum[int64])
			if !ok {
				return 0, false
			}

			for _, point := range sum.DataPoints {
				if attributesMatch(point.Attributes, expectedAttrs) {
					return point.Value, true
				}
			}
		}
	}

	return 0, false
}

func pluginInvocationHistogramCount(
	metrics metricdata.ResourceMetrics,
	expectedAttrs map[string]string,
) (uint64, bool) {
	for _, scopeMetrics := range metrics.ScopeMetrics {
		for _, metric := range scopeMetrics.Metrics {
			if metric.Name != pluginInvocationDurationMetricName {
				continue
			}

			histogram, ok := metric.Data.(metricdata.Histogram[float64])
			if !ok {
				return 0, false
			}

			for _, point := range histogram.DataPoints {
				if attributesMatch(point.Attributes, expectedAttrs) {
					return point.Count, true
				}
			}
		}
	}

	return 0, false
}

func attributesMatch(set attribute.Set, expected map[string]string) bool {
	for key, value := range expected {
		got, ok := set.Value(attribute.Key(key))
		if !ok || got.AsString() != value {
			return false
		}
	}

	return true
}
