package debugging_runtime

type PluginRuntimeNotifier interface {
	// on runtime connected
	OnRuntimeConnected(*RemotePluginRuntime)

	// on runtime disconnected
	OnRuntimeDisconnected(*RemotePluginRuntime)
}
