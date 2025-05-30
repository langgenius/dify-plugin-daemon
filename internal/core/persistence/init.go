package persistence

import (
	"github.com/langgenius/dify-cloud-kit/oss"

	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

var (
	persistence *Persistence
)

func InitPersistence(oss oss.OSS, config *app.Config) {
	persistence = &Persistence{
		storage:        NewWrapper(oss, config.PersistenceStoragePath),
		maxStorageSize: config.PersistenceStorageMaxSize,
	}

	log.Info("Persistence initialized")
}

func GetPersistence() *Persistence {
	return persistence
}
