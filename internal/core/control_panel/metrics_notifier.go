package controlpanel

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/debugging_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/metrics"
)

type MetricsNotifier struct{}

func NewMetricsNotifier() *MetricsNotifier {
	return &MetricsNotifier{}
}

// identifiableRuntime is an interface for types that can provide their identity
type identifiableRuntime interface {
	Identity() (plugin_entities.PluginUniqueIdentifier, error)
}

func (m *MetricsNotifier) OnLocalRuntimeStarting(pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier) {
	pluginID := pluginUniqueIdentifier.PluginID()
	metrics.PluginRuntimeStatus.WithLabelValues(
		pluginID,
		"local",
	).Set(0.5)
}

func (m *MetricsNotifier) OnLocalRuntimeReady(runtime *local_runtime.LocalPluginRuntime) {
	pluginID := pluginIDFromRuntime(runtime)
	metrics.PluginRuntimeStatus.WithLabelValues(
		pluginID,
		"local",
	).Set(1)
	metrics.ActivePluginsTotal.WithLabelValues("local").Inc()
}

func (m *MetricsNotifier) OnLocalRuntimeStartFailed(pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier, err error) {
	pluginID := pluginUniqueIdentifier.PluginID()
	metrics.PluginRuntimeStatus.WithLabelValues(
		pluginID,
		"local",
	).Set(0)
	metrics.PluginInstallationsTotal.WithLabelValues(
		pluginID,
		"failed",
	).Inc()
}

func (m *MetricsNotifier) OnLocalRuntimeStopped(runtime *local_runtime.LocalPluginRuntime) {
	pluginID := pluginIDFromRuntime(runtime)
	metrics.PluginRuntimeStatus.WithLabelValues(
		pluginID,
		"local",
	).Set(0)
	metrics.ActivePluginsTotal.WithLabelValues("local").Dec()
}

func (m *MetricsNotifier) OnLocalRuntimeStop(runtime *local_runtime.LocalPluginRuntime) {
	pluginID := pluginIDFromRuntime(runtime)
	metrics.PluginRuntimeStatus.WithLabelValues(
		pluginID,
		"local",
	).Set(0)
}

func (m *MetricsNotifier) OnLocalRuntimeScaleUp(runtime *local_runtime.LocalPluginRuntime, i int32) {
}

func (m *MetricsNotifier) OnLocalRuntimeScaleDown(runtime *local_runtime.LocalPluginRuntime, i int32) {
}

func (m *MetricsNotifier) OnLocalRuntimeInstanceLog(
	runtime *local_runtime.LocalPluginRuntime,
	instance *local_runtime.PluginInstance,
	event plugin_entities.PluginLogEvent,
) {
}

func (m *MetricsNotifier) OnDebuggingRuntimeConnected(runtime *debugging_runtime.RemotePluginRuntime) {
	pluginID := pluginIDFromRuntime(runtime)
	metrics.PluginRuntimeStatus.WithLabelValues(
		pluginID,
		"remote",
	).Set(1)
	metrics.ActivePluginsTotal.WithLabelValues("remote").Inc()
}

func (m *MetricsNotifier) OnDebuggingRuntimeDisconnected(runtime *debugging_runtime.RemotePluginRuntime) {
	pluginID := pluginIDFromRuntime(runtime)
	metrics.PluginRuntimeStatus.WithLabelValues(
		pluginID,
		"remote",
	).Set(0)
	metrics.ActivePluginsTotal.WithLabelValues("remote").Dec()
}

// pluginIDFromRuntime extracts the plugin ID from any runtime that implements identifiableRuntime
func pluginIDFromRuntime(runtime identifiableRuntime) string {
	if runtime == nil {
		return "unknown"
	}
	if identity, err := runtime.Identity(); err == nil {
		return identity.PluginID()
	}
	return "unknown"
}
