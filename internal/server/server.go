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
	// init routine pool
	routine.InitPool(config.RoutinePoolSize)

	// init db
	db.Init(config)

	// init process lifetime
	process.Init(config)

	// create manager
	manager := plugin_manager.NewManager(config)

	// create cluster
	app.cluster = cluster.NewCluster(config, manager)

	// register plugin lifetime event
	manager.AddPluginRegisterHandler(app.cluster.RegisterPlugin)

	// init manager
	manager.Init(config)

	// init persistence
	persistence.InitPersistence(config)

	// launch cluster
	app.cluster.Launch()

	// start http server
	app.server(config)

	// block
	select {}
}
