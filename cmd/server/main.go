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

	server.Run(&config)
}

func setDefault(config *app.Config) {
	setDefaultInt(&config.RoutinePoolSize, 1000)
	setDefaultInt(&config.DifyCallingPort, 5002)
}

func setDefaultInt[T constraints.Integer](value *T, defaultValue T) {
	if *value == 0 {
		*value = defaultValue
	}
}
