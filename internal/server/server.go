package server

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/process"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

func Run(config *app.Config) {
	// init routine pool
	routine.InitPool(config.RoutinePoolSize)

	// init db
	db.Init(config)

	// init process lifetime
	process.Init(config)

	// init plugin daemon
	plugin_manager.Init(config)

	// start http server
	server(config)
}
