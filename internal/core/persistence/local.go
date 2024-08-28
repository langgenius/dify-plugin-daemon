package persistence

import (
	"os"
	"path"
)

type LocalWrapper struct {
	path string
}

func NewLocalWrapper(path string) *LocalWrapper {
	return &LocalWrapper{
		path: path,
	}
}

func (l *LocalWrapper) getFilePath(tenant_id string, plugin_checksum string, key string) string {
	return path.Join(l.path, tenant_id, plugin_checksum, key)
}

func (l *LocalWrapper) Save(tenant_id string, plugin_checksum string, key string, data []byte) error {
	file_path := l.getFilePath(tenant_id, plugin_checksum, key)
	return os.WriteFile(file_path, data, 0644)
}

func (l *LocalWrapper) Load(tenant_id string, plugin_checksum string, key string) ([]byte, error) {
	file_path := l.getFilePath(tenant_id, plugin_checksum, key)
	return os.ReadFile(file_path)
}

func (l *LocalWrapper) Delete(tenant_id string, plugin_checksum string, key string) error {
	file_path := l.getFilePath(tenant_id, plugin_checksum, key)
	return os.Remove(file_path)
}
