package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "plugin_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "plugin_http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	PluginInvocationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "plugin_invocations_total",
			Help: "Total number of plugin invocations",
		},
		[]string{"plugin_id", "plugin_type", "runtime_type", "operation", "status"},
	)

	PluginInvocationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "plugin_invocation_duration_seconds",
			Help:    "Plugin invocation latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"plugin_id", "plugin_type", "runtime_type", "operation"},
	)

	PluginInvocationsActive = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "plugin_invocations_active",
			Help: "Number of active plugin invocations",
		},
		[]string{"plugin_id", "plugin_type", "runtime_type"},
	)

	PluginInstallationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "plugin_installations_total",
			Help: "Total number of plugin installations",
		},
		[]string{"plugin_id", "status"},
	)

	PluginInstallationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "plugin_installation_duration_seconds",
			Help:    "Plugin installation latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"plugin_id"},
	)

	PluginRuntimeStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "plugin_runtime_status",
			Help: "Current status of plugin runtime (1=active, 0.5=launching, 0=stopped)",
		},
		[]string{"plugin_id", "runtime_type"},
	)

	PluginRestartsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "plugin_restarts_total",
			Help: "Total number of plugin restarts",
		},
		[]string{"plugin_id", "runtime_type"},
	)

	ActivePluginsTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "active_plugins_total",
			Help: "Total number of active plugins",
		},
		[]string{"runtime_type"},
	)

	ActiveSessionsTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_sessions_total",
			Help: "Total number of active sessions",
		},
	)

	StorageOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "storage_operations_total",
			Help: "Total number of storage operations",
		},
		[]string{"operation", "storage_type", "status"},
	)

	StorageOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "storage_operation_duration_seconds",
			Help:    "Storage operation latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "storage_type"},
	)
)
