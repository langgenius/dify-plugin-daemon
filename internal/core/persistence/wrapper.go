package persistence

import (
	"path"

	"github.com/langgenius/dify-plugin-daemon/internal/oss"
)

type wrapper struct {
	oss                      oss.OSS
	persistence_storage_path string
}

func NewWrapper(oss oss.OSS, persistence_storage_path string) *wrapper {
	return &wrapper{
		oss:                      oss,
		persistence_storage_path: persistence_storage_path,
	}
}

func (s *wrapper) getFilePath(tenant_id string, plugin_checksum string, key string) string {
	return path.Join(s.persistence_storage_path, tenant_id, plugin_checksum, key)
}

func (s *wrapper) Save(tenant_id string, plugin_checksum string, key string, data []byte) error {
	// save to s3
	file_path := s.getFilePath(tenant_id, plugin_checksum, key)
	return s.oss.Save(file_path, data)
}

func (s *wrapper) Load(tenant_id string, plugin_checksum string, key string) ([]byte, error) {
	// load from s3
	file_path := s.getFilePath(tenant_id, plugin_checksum, key)
	return s.oss.Load(file_path)
}

func (s *wrapper) Delete(tenant_id string, plugin_checksum string, key string) error {
	file_path := s.getFilePath(tenant_id, plugin_checksum, key)
	return s.oss.Delete(file_path)
}

func (s *wrapper) StateSize(tenant_id string, plugin_checksum string, key string) (int64, error) {
	file_path := s.getFilePath(tenant_id, plugin_checksum, key)
	state, err := s.oss.State(file_path)
	if err != nil {
		return 0, err
	}

	return state.Size, nil
}
