package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Issue #812: operational metrics for plugin invoke saturation and failures.
var (
	PluginDaemonInvokesInFlight = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "plugin_daemon_invokes_in_flight",
			Help: "Number of plugin invokes currently queued or executing",
		},
		[]string{"plugin"},
	)

	PluginDaemonInvokeDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "plugin_daemon_invoke_duration_seconds",
			Help:    "Plugin invoke latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"plugin", "outcome"},
	)

	PluginDaemonPluginProcesses = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "plugin_daemon_plugin_processes",
			Help: "Number of running plugin worker processes for a plugin",
		},
		[]string{"plugin"},
	)

	PluginDaemonInvokeErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "plugin_daemon_invoke_errors_total",
			Help: "Total number of failed plugin invokes by error class",
		},
		[]string{"plugin", "error_class"},
	)
)

// InvokeOutcome classifies how an invoke finished for metrics labels.
const (
	InvokeOutcomeSuccess          = "success"
	InvokeOutcomeError            = "error"
	InvokeOutcomeTimeout          = "timeout"
	InvokeOutcomeClientDisconnect = "client_disconnect"
	InvokeOutcomeSessionError     = "session_error"
)

// InvokeErrorClass is a bounded set of failure modes for counters.
const (
	InvokeErrorClassPlugin = "plugin_error"
	InvokeErrorClassTimeout = "timeout"
	InvokeErrorClassCanceled = "canceled"
	InvokeErrorClassDaemon   = "daemon_error"
	InvokeErrorClassOther    = "other"
)

func invokeErrorClassForOutcome(outcome string) (string, bool) {
	switch outcome {
	case InvokeOutcomeSuccess:
		return "", false
	case InvokeOutcomeTimeout:
		return InvokeErrorClassTimeout, true
	case InvokeOutcomeClientDisconnect:
		return InvokeErrorClassCanceled, true
	case InvokeOutcomeSessionError:
		return InvokeErrorClassDaemon, true
	case InvokeOutcomeError:
		return InvokeErrorClassPlugin, true
	default:
		return InvokeErrorClassOther, true
	}
}

func RecordPluginDaemonInvokeStart(pluginID string) {
	PluginDaemonInvokesInFlight.WithLabelValues(pluginID).Inc()
}

// RecordPluginDaemonInvokeComplete records the end of an invoke that incremented in-flight.
func RecordPluginDaemonInvokeComplete(pluginID, outcome string, durationSeconds float64) {
	recordPluginDaemonInvokeEnd(pluginID, outcome, durationSeconds, true)
}

// RecordPluginDaemonInvokeFailed records a failed invoke that never entered the in-flight gauge.
func RecordPluginDaemonInvokeFailed(pluginID, outcome string, durationSeconds float64) {
	recordPluginDaemonInvokeEnd(pluginID, outcome, durationSeconds, false)
}

func recordPluginDaemonInvokeEnd(
	pluginID, outcome string,
	durationSeconds float64,
	decrementInFlight bool,
) {
	if decrementInFlight {
		PluginDaemonInvokesInFlight.WithLabelValues(pluginID).Dec()
	}
	PluginDaemonInvokeDurationSeconds.WithLabelValues(pluginID, outcome).Observe(durationSeconds)
	if errorClass, record := invokeErrorClassForOutcome(outcome); record {
		PluginDaemonInvokeErrorsTotal.WithLabelValues(pluginID, errorClass).Inc()
	}
}

func SetPluginDaemonPluginProcesses(pluginID string, count float64) {
	PluginDaemonPluginProcesses.WithLabelValues(pluginID).Set(count)
}
