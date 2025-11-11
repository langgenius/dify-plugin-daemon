package local_runtime

type PluginRuntimeNotifier interface {
	// on instance starting
	OnInstanceStarting(*PluginInstance)

	// on instance ready
	OnInstanceReady(*PluginInstance)

	// on instance failed
	OnInstanceLaunchFailed(*PluginInstance, error)

	// on instance shutdown
	OnInstanceShutdown(*PluginInstance)

	// on instance scale down failed
	OnInstanceScaleDownFailed(error)

	// on runtime stop schedule
	OnRuntimeStopSchedule()

	// on runtime close
	OnRuntimeClose()
}
