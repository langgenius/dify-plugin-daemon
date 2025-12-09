package local_runtime

type PluginRuntimeNotifier interface {
	// on instance starting
	OnInstanceStarting()

	// on instance ready
	OnInstanceReady(*PluginInstance)

	// on instance failed
	OnInstanceLaunchFailed(*PluginInstance, error)

	// on instance shutdown
	OnInstanceShutdown(*PluginInstance)

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
