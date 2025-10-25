package debugging_runtime

type PluginRuntimeNotifier interface {
	// on runtime connected
	OnRuntimeConnected(*RemotePluginRuntime) error

	// on runtime disconnected
	OnRuntimeDisconnected(*RemotePluginRuntime)
}
