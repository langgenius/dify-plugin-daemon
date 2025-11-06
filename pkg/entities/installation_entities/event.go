package installation_entities

// Streaming events for plugin installation
type PluginInstallEvent string

const (
	PluginInstallEventInfo  PluginInstallEvent = "info"
	PluginInstallEventDone  PluginInstallEvent = "done"
	PluginInstallEventError PluginInstallEvent = "error"
)

type PluginInstallResponse struct {
	Event PluginInstallEvent `json:"event"`
	Data  string             `json:"data"`
}
