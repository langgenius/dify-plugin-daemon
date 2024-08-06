package plugin_daemon

import "github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/access_types"

func getBasicPluginAccessMap(
	user_id string,
	access_type access_types.PluginAccessType,
	action access_types.PluginAccessAction,
) map[string]any {
	return map[string]any{
		"user_id": user_id,
		"type":    access_type,
		"action":  action,
	}
}
