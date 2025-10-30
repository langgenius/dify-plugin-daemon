package controlpanel

import serverless "github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/serverless_connector"

// InstallLocalPluginResponse is the response type for local plugin installation
type InstallLocalPluginResponse struct {
}

// InstallServerlessPluginResponse is the response type for serverless plugin installation
type InstallServerlessPluginResponse = serverless.LaunchFunctionResponse
