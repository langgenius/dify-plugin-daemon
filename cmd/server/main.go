package main

import (
	"context"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/langgenius/dify-plugin-daemon/internal/server"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

func main() {
	var config app.Config

	// load env
	err := godotenv.Load()
	if err != nil {
		log.Panic("failed to load .env file", "error", err)
	}

	err = envconfig.Process("", &config)
	if err != nil {
		log.Panic("error processing environment variables", "error", err)
	}

	config.SetDefault()

	logCloser, err := log.Init(config.LogOutputFormat == "json", config.LogFile)
	if err != nil {
		log.Panic("failed to init logger", "error", err)
	}
	if logCloser != nil {
		defer func() {
			if err := logCloser.Close(); err != nil {
				log.Error("failed to close log file", "error", err)
			}
		}()
	}
	defer log.RecoverAndExit()

	if err = config.Validate(); err != nil {
		log.Panic("invalid configuration", "error", err)
	}

	// Initialize OpenTelemetry if enabled
	if config.EnableOtel {
		shutdown, err := server.InitTelemetry(&config)
		if err != nil {
			log.Panic("failed to init OpenTelemetry", "error", err)
		} else {
			defer shutdown(context.Background())
		}
	}

	(&server.App{}).Run(&config)
}
