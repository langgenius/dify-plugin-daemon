package controlpanel

import serverless "github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/serverless_connector"

type LocalPluginLaunchEvent string

const (
	Error LocalPluginLaunchEvent = "error"
	Info  LocalPluginLaunchEvent = "info"
	Done  LocalPluginLaunchEvent = "done"
)

// InstallLocalPluginResponse is the response type for local plugin installation
type InstallLocalPluginResponse struct {
	Event   LocalPluginLaunchEvent `json:"event"`
	Message string                 `json:"message"`
}

// InstallServerlessPluginResponse is the response type for serverless plugin installation
type InstallServerlessPluginResponse = serverless.LaunchFunctionResponse
