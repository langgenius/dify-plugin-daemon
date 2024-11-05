package persistence

import (
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

var (
	persistence *Persistence
)

func InitPersistence(config *app.Config) {
	if config.PersistenceStorageType == "s3" {
		s3, err := NewS3Wrapper(
			config.PersistenceStorageS3Region,
			config.PersistenceStorageS3AccessKey,
			config.PersistenceStorageS3SecretKey,
			config.PersistenceStorageS3Bucket,
		)
		if err != nil {
			log.Panic("Failed to initialize S3 wrapper: %v", err)
		}

		persistence = &Persistence{
			storage:          s3,
			max_storage_size: config.PersistenceStorageMaxSize,
		}
	} else if config.PersistenceStorageType == "local" {
		persistence = &Persistence{
			storage:          NewLocalWrapper(config.PersistenceStorageLocalPath),
			max_storage_size: config.PersistenceStorageMaxSize,
		}
	} else {
		log.Panic("Invalid persistence storage type: %s", config.PersistenceStorageType)
	}

	log.Info("Persistence initialized")
}

func GetPersistence() *Persistence {
	return persistence
}
