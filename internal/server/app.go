package server

import (
	"github.com/langgenius/dify-plugin-daemon/internal/cluster"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
)

type App struct {
	plugin_manager *plugin_manager.PluginManager
	cluster        *cluster.Cluster
}
