package plugin_manager

import (
	"errors"
)

var (
	ErrPluginPackageNotFound = errors.New("plugin package not found, please upload it firstly")
)
