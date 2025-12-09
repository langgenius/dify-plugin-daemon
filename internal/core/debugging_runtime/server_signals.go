package debugging_runtime

type ServerShutdownReason string

const (
	SERVER_SHUTDOWN_REASON_EXIT  ServerShutdownReason = "exit"
	SERVER_SHUTDOWN_REASON_ERROR ServerShutdownReason = "error"
)

type PluginRuntimeNotifier interface {
	// on runtime connected
	OnRuntimeConnected(*RemotePluginRuntime) error

	// on runtime disconnected
	OnRuntimeDisconnected(*RemotePluginRuntime)

	// on server shutdown
	OnServerShutdown(reason ServerShutdownReason)
}
