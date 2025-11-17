package helper

import "strings"

func PluginInstallationCacheKey(pluginId, tenantId string) string {
	return strings.Join(
		[]string{
			"plugin_id",
			pluginId,
			"tenant_id",
			tenantId,
		},
		":",
	)
}

func EndpointCacheKey(hookId string) string {
	return strings.Join(
		[]string{
			"hook_id",
			hookId,
		},
		":",
	)
}