package controlpanel

import "errors"

var (
	ErrorPluginAlreadyLaunched    = errors.New("plugin already launched")
	ErrLocalPluginRuntimeNotFound = errors.New("local plugin runtime not found")
)
