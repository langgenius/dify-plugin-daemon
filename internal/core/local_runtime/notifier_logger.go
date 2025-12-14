package local_runtime

import (
	"strings"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

type NotifierLogger struct {
}

func (n *NotifierLogger) OnInstanceStarting() {
	// Nop
}

func (n *NotifierLogger) OnInstanceReady(instance *PluginInstance) {
	// notify terminal
	log.Info("plugin %s: instance %s ready", instance.pluginUniqueIdentifier, instance.instanceId[:8])
}

func (n *NotifierLogger) OnInstanceLaunchFailed(instance *PluginInstance, err error) {
	log.Error("plugin %s: instance %s failed: %s", instance.pluginUniqueIdentifier, instance.instanceId[:8], err.Error())
}

func (n *NotifierLogger) OnInstanceShutdown(instance *PluginInstance) {
	// notify terminal
	log.Warn("plugin %s: instance %s has been shutdown", instance.pluginUniqueIdentifier, instance.instanceId[:8])
}

func (n *NotifierLogger) OnInstanceHeartbeat(instance *PluginInstance) {
	// Nop
}

func (n *NotifierLogger) OnInstanceLog(instance *PluginInstance, event plugin_entities.PluginLogEvent) {
	// notify terminal
	instanceID := instance.instanceId[:8]

	switch strings.ToLower(event.Level) {
	case "debug":
		log.Debug("plugin %s: instance %s log: %s", instance.pluginUniqueIdentifier, instanceID, event.Message)
	case "warn", "warning":
		log.Warn("plugin %s: instance %s log: %s", instance.pluginUniqueIdentifier, instanceID, event.Message)
	case "error", "fatal", "critical":
		log.Error("plugin %s: instance %s log: %s", instance.pluginUniqueIdentifier, instanceID, event.Message)
	default:
		log.Info("plugin %s: instance %s log: %s", instance.pluginUniqueIdentifier, instanceID, event.Message)
	}
}

func (n *NotifierLogger) OnInstanceErrorLog(instance *PluginInstance, err error) {
	// notify terminal
	log.Error(
		"plugin %s: instance %s get an error message: %s",
		instance.pluginUniqueIdentifier,
		instance.instanceId[:8],
		err.Error(),
	)
}

func (n *NotifierLogger) OnInstanceWarningLog(instance *PluginInstance, message string) {
	// notify terminal
	log.Warn(
		"plugin %s: instance %s get a warning message: %s",
		instance.pluginUniqueIdentifier,
		instance.instanceId[:8],
		message,
	)
}

func (n *NotifierLogger) OnInstanceStdout(instance *PluginInstance, data []byte) {
	// nop
}

func (n *NotifierLogger) OnInstanceStderr(instance *PluginInstance, data []byte) {
	// nop
}
