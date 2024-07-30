package server

import (
	"github.com/langgenius/dify-plugin-daemon/internal/cluster"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/process"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

func (a *App) Run(config *app.Config) {
	a.cluster = cluster.NewCluster(config)
	plugin_manager.InitGlobalPluginManager(a.cluster)

	// init routine pool
	routine.InitPool(config.RoutinePoolSize)

	// init db
	db.Init(config)

	// init process lifetime
	process.Init(config)

	// init plugin daemon
	a.plugin_manager.Init(config)

	// launch cluster
	a.cluster.Launch(config)

	// start http server
	server(config)
}
