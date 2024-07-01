package server

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
)

func Run(config *app.Config) {
	// init plugin daemon
	plugin_manager.Init(config)
}
