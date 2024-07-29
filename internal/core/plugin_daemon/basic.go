package plugin_daemon

import "github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/backwards_invocation"

func getBasicPluginAccessMap(user_id string, access_type backwards_invocation.PluginAccessType, action backwards_invocation.PluginAccessAction) map[string]any {
	return map[string]any{
		"user_id": user_id,
		"type":    access_type,
		"action":  action,
	}
}
