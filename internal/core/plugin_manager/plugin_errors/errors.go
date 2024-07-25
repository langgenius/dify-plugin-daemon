package plugin_errors

import "errors"

var (
	ErrPluginNotActive = errors.New("plugin is not active, does not respond to heartbeat in 20 seconds")
)
