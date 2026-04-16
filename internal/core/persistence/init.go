package persistence

import (
	"path"
	"strings"

	"github.com/langgenius/dify-cloud-kit/oss"

	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

var (
	persistence *Persistence
)

func InitPersistence(oss oss.OSS, config *app.Config) {
	storagePath := config.PersistenceStoragePath
	if prefix := strings.Trim(config.StoragePathPrefix, "/"); prefix != "" {
		for _, seg := range strings.Split(prefix, "/") {
			if seg == ".." {
				log.Panic("STORAGE_PATH_PREFIX must not contain '..'")
			}
		}
		storagePath = path.Join(prefix, storagePath)
	}

	persistence = &Persistence{
		storage:        NewWrapper(oss, storagePath),
		maxStorageSize: config.PersistenceStorageMaxSize,
	}

	log.Info("persistence initialized")
}

func GetPersistence() *Persistence {
	return persistence
}
