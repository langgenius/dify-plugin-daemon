package controlpanel

import (
	"reflect"

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
	metrics.SetPluginDaemonPluginProcesses(pluginID, 0)
}

func (m *MetricsNotifier) OnLocalRuntimeStop(runtime *local_runtime.LocalPluginRuntime) {
	pluginID := pluginIDFromRuntime(runtime)
	metrics.PluginRuntimeStatus.WithLabelValues(
		pluginID,
		"local",
	).Set(0)
}

func (m *MetricsNotifier) OnLocalRuntimeScaleUp(runtime *local_runtime.LocalPluginRuntime, i int32) {
	pluginID := pluginIDFromRuntime(runtime)
	metrics.SetPluginDaemonPluginProcesses(pluginID, float64(i))
}

func (m *MetricsNotifier) OnLocalRuntimeScaleDown(runtime *local_runtime.LocalPluginRuntime, i int32) {
	pluginID := pluginIDFromRuntime(runtime)
	metrics.SetPluginDaemonPluginProcesses(pluginID, float64(i))
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
	metrics.SetPluginDaemonPluginProcesses(pluginID, 1)
}

func (m *MetricsNotifier) OnDebuggingRuntimeDisconnected(runtime *debugging_runtime.RemotePluginRuntime) {
	pluginID := pluginIDFromRuntime(runtime)
	metrics.PluginRuntimeStatus.WithLabelValues(
		pluginID,
		"remote",
	).Set(0)
	metrics.ActivePluginsTotal.WithLabelValues("remote").Dec()
	metrics.SetPluginDaemonPluginProcesses(pluginID, 0)
}

// pluginIDFromRuntime extracts the plugin ID from any runtime that implements identifiableRuntime
func pluginIDFromRuntime(runtime identifiableRuntime) string {
	if isNilIdentifiableRuntime(runtime) {
		return "unknown"
	}
	if identity, err := runtime.Identity(); err == nil {
		return identity.PluginID()
	}
	return "unknown"
}

func isNilIdentifiableRuntime(runtime identifiableRuntime) bool {
	if runtime == nil {
		return true
	}
	value := reflect.ValueOf(runtime)
	switch value.Kind() {
	case reflect.Pointer, reflect.Interface, reflect.Map, reflect.Slice, reflect.Func, reflect.Chan:
		return value.IsNil()
	default:
		return false
	}
}
