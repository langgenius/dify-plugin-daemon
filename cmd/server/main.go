package main

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/langgenius/dify-plugin-daemon/internal/server"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

func main() {
	var config app.Config

	// load env
	godotenv.Load()

	err := envconfig.Process("", &config)
	if err != nil {
		log.Panic("error processing environment variables", "error", err)
	}

	config.SetDefault()

	log.Init(config.LogOutputFormat == "json")
	defer log.RecoverAndExit()

	if err := config.Validate(); err != nil {
		log.Panic("invalid configuration", "error", err)
	}

	(&server.App{}).Run(&config)
}
