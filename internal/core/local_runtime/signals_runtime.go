package local_runtime

import "github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"

type PluginRuntimeNotifier interface {
	// on instance starting
	OnInstanceStarting()

	// on instance ready
	OnInstanceReady(*PluginInstance)

	// on instance failed
	OnInstanceLaunchFailed(*PluginInstance, error)

	// on instance shutdown
	OnInstanceShutdown(*PluginInstance)

	// on instance log
	OnInstanceLog(*PluginInstance, plugin_entities.PluginLogEvent)

	// on instance scale up
	OnInstanceScaleUp(int32)

	// on instance scale down
	OnInstanceScaleDown(int32)

	// on instance scale down failed
	OnInstanceScaleDownFailed(error)

	// on runtime stop schedule
	OnRuntimeStopSchedule()

	// on runtime close
	OnRuntimeClose()
}
