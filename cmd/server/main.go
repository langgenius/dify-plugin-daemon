package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/langgenius/dify-plugin-daemon/internal/server"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

func main() {
	var config app.Config

	err := loadDotEnv()
	if err != nil {
		log.Panic("failed to load .env file", "error", err)
	}

	err = envconfig.Process("", &config)
	if err != nil {
		log.Panic("error processing environment variables", "error", err)
	}

	config.SetDefault()

	log.Init(config.LogOutputFormat == "json")
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

func loadDotEnv() error {
	const (
		dotEnvModeEnvVar = "DIFY_DOTENV_MODE"
		dotEnvFileEnvVar = "DIFY_ENV_FILE"

		dotEnvModeOptional = "optional"
		dotEnvModeRequire  = "require"
		dotEnvModeDisabled = "disabled"
	)

	dotEnvMode := strings.ToLower(strings.TrimSpace(os.Getenv(dotEnvModeEnvVar)))
	if dotEnvMode == "" {
		dotEnvMode = dotEnvModeOptional
	}

	switch dotEnvMode {
	case dotEnvModeOptional, dotEnvModeRequire, dotEnvModeDisabled:
	default:
		return fmt.Errorf("invalid %s: %s (valid options: %s, %s, %s)", dotEnvModeEnvVar, dotEnvMode, dotEnvModeOptional, dotEnvModeRequire, dotEnvModeDisabled)
	}

	if dotEnvMode == dotEnvModeDisabled {
		return nil
	}

	dotEnvFilePath := strings.TrimSpace(os.Getenv(dotEnvFileEnvVar))
	if dotEnvFilePath == "" {
		dotEnvFilePath = ".env"
	}

	fileInfo, err := os.Stat(dotEnvFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			if dotEnvMode == dotEnvModeRequire {
				return fmt.Errorf("required .env file not found: %s", dotEnvFilePath)
			}
			return nil
		}

		return fmt.Errorf("failed to stat .env file %s: %w", dotEnvFilePath, err)
	}

	if fileInfo.IsDir() {
		return fmt.Errorf(".env path is a directory: %s", dotEnvFilePath)
	}

	dotEnvValues, err := godotenv.Read(dotEnvFilePath)
	if err != nil {
		return fmt.Errorf("invalid .env file %s: %w", dotEnvFilePath, err)
	}

	for key, value := range dotEnvValues {
		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		err = os.Setenv(key, value)
		if err != nil {
			return fmt.Errorf("failed to set env var from .env (%s): %w", key, err)
		}
	}

	return nil
}
