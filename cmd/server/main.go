package main

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/langgenius/dify-plugin-daemon/internal/server"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"golang.org/x/exp/constraints"
)

func main() {
	var config app.Config

	err := godotenv.Load()
	if err != nil {
		log.Panic("Error loading .env file")
	}

	err = envconfig.Process("", &config)
	if err != nil {
		log.Panic("Error processing environment variables")
	}

	setDefault(&config)

	if err := config.Validate(); err != nil {
		log.Panic("Invalid configuration: %s", err.Error())
	}

	server.Run(&config)
}

func setDefault(config *app.Config) {
	setDefaultInt(&config.ServerPort, 5002)
	setDefaultInt(&config.RoutinePoolSize, 1000)
	setDefaultInt(&config.LifetimeCollectionGCInterval, 60)
	setDefaultInt(&config.LifetimeCollectionHeartbeatInterval, 5)
	setDefaultInt(&config.LifetimeStateGCInterval, 300)
	setDefaultInt(&config.DifyInvocationConnectionIdleTimeout, 120)
	setDefaultInt(&config.PluginRemoteInstallServerEventLoopNums, 8)
	setDefaultInt(&config.PluginRemoteInstallingMaxConn, 128)
	settDefaultBool(&config.PluginRemoteInstallingEnabled, true)

	settDefaultString(&config.ProcessCachingPath, "/tmp/dify-plugin-daemon-subprocesses")
}

func setDefaultInt[T constraints.Integer](value *T, defaultValue T) {
	if *value == 0 {
		*value = defaultValue
	}
}

func settDefaultString(value *string, defaultValue string) {
	if *value == "" {
		*value = defaultValue
	}
}

func settDefaultBool(value *bool, defaultValue bool) {
	if !*value {
		*value = defaultValue
	}
}
