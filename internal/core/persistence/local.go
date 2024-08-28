package persistence

import (
	"os"
	"path"
)

type LocalWrapper struct {
	path string
}

func NewLocalWrapper(path string) *LocalWrapper {
	// check if the path exists, create it if not
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0755)
	}

	return &LocalWrapper{
		path: path,
	}
}

func (l *LocalWrapper) getFilePath(tenant_id string, plugin_checksum string, key string) string {
	return path.Join(l.path, tenant_id, plugin_checksum, key)
}

func (l *LocalWrapper) Save(tenant_id string, plugin_checksum string, key string, data []byte) error {
	// create the directory if it doesn't exist
	dir := l.getFilePath(tenant_id, plugin_checksum, "")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

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
