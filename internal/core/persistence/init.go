package persistence

import (
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

func InitPersistence(config *app.Config) *Persistence {
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

		return &Persistence{
			storage: s3,
		}
	}

	return &Persistence{
		storage: NewLocalWrapper(),
	}
}
