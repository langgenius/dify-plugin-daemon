package controlpanel

import (
	serverless "github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/serverless_connector"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/installation_entities"
)

type LocalPluginLaunchEvent = installation_entities.PluginInstallEvent

// InstallLocalPluginResponse is the response type for local plugin installation
type InstallLocalPluginResponse struct {
	Event   LocalPluginLaunchEvent `json:"event"`
	Message string                 `json:"message"`
}

// InstallServerlessPluginResponse is the response type for serverless plugin installation
type InstallServerlessPluginResponse = serverless.LaunchFunctionResponse
