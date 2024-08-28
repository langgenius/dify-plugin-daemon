package server

import (
	"github.com/langgenius/dify-plugin-daemon/internal/cluster"
	"github.com/langgenius/dify-plugin-daemon/internal/core/persistence"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/process"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

func (a *App) Run(config *app.Config) {
	a.cluster = cluster.NewCluster(config)

	// init routine pool
	routine.InitPool(config.RoutinePoolSize)

	// init db
	db.Init(config)

	// init process lifetime
	process.Init(config)

	// init plugin daemon
	plugin_manager.InitGlobalPluginManager(a.cluster, config)

	// init persistence
	persistence.InitPersistence(config)

	// launch cluster
	a.cluster.Launch()

	// start http server
	a.server(config)

	// block
	select {}
}
