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

func (app *App) Run(config *app.Config) {
	app.cluster = cluster.NewCluster(config)

	// init routine pool
	routine.InitPool(config.RoutinePoolSize)

	// init db
	db.Init(config)

	// init process lifetime
	process.Init(config)

	// init plugin daemon
	plugin_manager.InitGlobalPluginManager(app.cluster, config)

	// init persistence
	persistence.InitPersistence(config)

	// launch cluster
	app.cluster.Launch()

	// start http server
	app.server(config)

	// block
	select {}
}
