package metrics

import (
	"testing"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetPluginDaemonInvokeMetrics(t *testing.T) {
	t.Helper()
	PluginDaemonInvokesInFlight.Reset()
	PluginDaemonInvokeDurationSeconds.Reset()
	PluginDaemonInvokeErrorsTotal.Reset()
}

func inFlightValue(t *testing.T, pluginID string) float64 {
	t.Helper()
	metric, err := PluginDaemonInvokesInFlight.GetMetricWithLabelValues(pluginID)
	require.NoError(t, err)
	var dtoMetric dto.Metric
	require.NoError(t, metric.Write(&dtoMetric))
	return dtoMetric.GetGauge().GetValue()
}

func TestRecordPluginDaemonInvokeLifecycle(t *testing.T) {
	resetPluginDaemonInvokeMetrics(t)

	RecordPluginDaemonInvokeStart("author/plugin")
	assert.Equal(t, 1.0, inFlightValue(t, "author/plugin"))

	RecordPluginDaemonInvokeComplete("author/plugin", InvokeOutcomeSuccess, 0.5)
	assert.Equal(t, 0.0, inFlightValue(t, "author/plugin"))

	observer, err := PluginDaemonInvokeDurationSeconds.GetMetricWithLabelValues(
		"author/plugin",
		InvokeOutcomeSuccess,
	)
	require.NoError(t, err)
	metric, ok := observer.(prometheus.Metric)
	require.True(t, ok)
	var hist dto.Metric
	require.NoError(t, metric.Write(&hist))
	assert.Equal(t, uint64(1), hist.GetHistogram().GetSampleCount())
}

func TestRecordPluginDaemonInvokeOutcomes(t *testing.T) {
	tests := []struct {
		name           string
		outcome        string
		decrement      bool
		wantErrorClass string
		wantInFlight   float64
	}{
		{
			name:         "success",
			outcome:      InvokeOutcomeSuccess,
			decrement:    true,
			wantInFlight: 0,
		},
		{
			name:           "plugin error",
			outcome:        InvokeOutcomeError,
			decrement:      true,
			wantErrorClass: InvokeErrorClassPlugin,
			wantInFlight:   0,
		},
		{
			name:           "timeout",
			outcome:        InvokeOutcomeTimeout,
			decrement:      true,
			wantErrorClass: InvokeErrorClassTimeout,
			wantInFlight:   0,
		},
		{
			name:           "client disconnect",
			outcome:        InvokeOutcomeClientDisconnect,
			decrement:      true,
			wantErrorClass: InvokeErrorClassCanceled,
			wantInFlight:   0,
		},
		{
			name:           "session error without in-flight",
			outcome:        InvokeOutcomeSessionError,
			decrement:      false,
			wantErrorClass: InvokeErrorClassDaemon,
			wantInFlight:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetPluginDaemonInvokeMetrics(t)
			pluginID := "author/plugin"

			if tt.decrement {
				RecordPluginDaemonInvokeStart(pluginID)
				RecordPluginDaemonInvokeComplete(pluginID, tt.outcome, 0.1)
			} else {
				RecordPluginDaemonInvokeFailed(pluginID, tt.outcome, 0.1)
			}

			assert.Equal(t, tt.wantInFlight, inFlightValue(t, pluginID))

			if tt.wantErrorClass == "" {
				return
			}

			counter, err := PluginDaemonInvokeErrorsTotal.GetMetricWithLabelValues(
				pluginID,
				tt.wantErrorClass,
			)
			require.NoError(t, err)
			var dtoMetric dto.Metric
			require.NoError(t, counter.Write(&dtoMetric))
			assert.Equal(t, 1.0, dtoMetric.GetCounter().GetValue())
		})
	}
}

func TestSetPluginDaemonPluginProcesses(t *testing.T) {
	PluginDaemonPluginProcesses.Reset()
	SetPluginDaemonPluginProcesses("author/plugin", 3)

	gauge, err := PluginDaemonPluginProcesses.GetMetricWithLabelValues("author/plugin")
	require.NoError(t, err)
	var dtoMetric dto.Metric
	require.NoError(t, gauge.Write(&dtoMetric))
	assert.Equal(t, 3.0, dtoMetric.GetGauge().GetValue())
}
