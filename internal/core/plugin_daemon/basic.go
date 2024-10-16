package plugin_daemon

import "github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/access_types"

func getBasicPluginAccessMap(
	user_id string,
	access_type access_types.PluginAccessType,
	action access_types.PluginAccessAction,
	conversation_id *string,
	message_id *string,
	app_id *string,
	endpoint_id *string,
) map[string]any {
	return map[string]any{
		"user_id":         user_id,
		"type":            access_type,
		"action":          action,
		"conversation_id": conversation_id,
		"message_id":      message_id,
		"app_id":          app_id,
		"endpoint_id":     endpoint_id,
	}
}
