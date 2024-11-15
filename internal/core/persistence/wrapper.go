package persistence

import (
	"path"

	"github.com/langgenius/dify-plugin-daemon/internal/oss"
)

type wrapper struct {
	oss                    oss.OSS
	persistenceStoragePath string
}

func NewWrapper(oss oss.OSS, persistenceStoragePath string) *wrapper {
	return &wrapper{
		oss:                    oss,
		persistenceStoragePath: persistenceStoragePath,
	}
}

func (s *wrapper) getFilePath(tenant_id string, plugin_checksum string, key string) string {
	return path.Join(s.persistenceStoragePath, tenant_id, plugin_checksum, key)
}

func (s *wrapper) Save(tenant_id string, plugin_checksum string, key string, data []byte) error {
	// save to s3
	filePath := s.getFilePath(tenant_id, plugin_checksum, key)
	return s.oss.Save(filePath, data)
}

func (s *wrapper) Load(tenant_id string, plugin_checksum string, key string) ([]byte, error) {
	// load from s3
	filePath := s.getFilePath(tenant_id, plugin_checksum, key)
	return s.oss.Load(filePath)
}

func (s *wrapper) Delete(tenant_id string, plugin_checksum string, key string) error {
	filePath := s.getFilePath(tenant_id, plugin_checksum, key)
	return s.oss.Delete(filePath)
}

func (s *wrapper) StateSize(tenant_id string, plugin_checksum string, key string) (int64, error) {
	filePath := s.getFilePath(tenant_id, plugin_checksum, key)
	state, err := s.oss.State(filePath)
	if err != nil {
		return 0, err
	}

	return state.Size, nil
}