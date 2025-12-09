package plugin_manager

import (
	"github.com/langgenius/dify-plugin-daemon/internal/cluster"
	"github.com/langgenius/dify-plugin-daemon/internal/core/debugging_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

// implement cluster.ClusterTunnel for interface `controlpanel.ControlPanelNotifier`
type ClusterTunnel struct {
	cluster *cluster.Cluster
}

func (t *ClusterTunnel) OnDebuggingRuntimeConnected(
	runtime *debugging_runtime.RemotePluginRuntime,
) {
	// register the plugin to the cluster
	if err := t.cluster.RegisterPlugin(runtime); err != nil {
		log.Error("failed to register plugin: %s", err.Error())
	}
}

func (t *ClusterTunnel) OnDebuggingRuntimeDisconnected(
	runtime *debugging_runtime.RemotePluginRuntime,
) {
	// unregister the plugin from the cluster
	if err := t.cluster.UnregisterPlugin(runtime); err != nil {
		log.Error("failed to unregister plugin: %s", err.Error())
	}
}

func (t *ClusterTunnel) OnLocalRuntimeReady(
	runtime *local_runtime.LocalPluginRuntime,
) {
	// register the plugin to the cluster
	if err := t.cluster.RegisterPlugin(runtime); err != nil {
		log.Error("failed to register plugin: %s", err.Error())
	}
}

func (t *ClusterTunnel) OnLocalRuntimeStartFailed(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	err error,
) {
	// NOP
}

func (t *ClusterTunnel) OnLocalRuntimeStarting(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) {
	// NOP
}

func (t *ClusterTunnel) OnLocalRuntimeStop(
	runtime *local_runtime.LocalPluginRuntime,
) {
	// unregister the plugin from the cluster
	if err := t.cluster.UnregisterPlugin(runtime); err != nil {
		log.Error("failed to unregister plugin: %s", err.Error())
	}
}

func (t *ClusterTunnel) OnLocalRuntimeStopped(
	pluginUniqueIdentifier *local_runtime.LocalPluginRuntime,
) {
	// NOP
}
