package local

import (
	"os"
	"path/filepath"

	"github.com/langgenius/dify-plugin-daemon/internal/oss"
)

type LocalStorage struct {
	root string
}

func NewLocalStorage(root string) oss.OSS {
	return &LocalStorage{root: root}
}

func (l *LocalStorage) Save(key string, data []byte) error {
	path := filepath.Join(l.root, key)

	return os.WriteFile(path, data, 0o644)
}

func (l *LocalStorage) Load(key string) ([]byte, error) {
	path := filepath.Join(l.root, key)

	return os.ReadFile(path)
}

func (l *LocalStorage) Exists(key string) (bool, error) {
	path := filepath.Join(l.root, key)

	_, err := os.Stat(path)
	return err == nil, nil
}

func (l *LocalStorage) State(key string) (oss.OSSState, error) {
	path := filepath.Join(l.root, key)

	info, err := os.Stat(path)
	if err != nil {
		return oss.OSSState{}, err
	}

	return oss.OSSState{Size: info.Size(), LastModified: info.ModTime()}, nil
}

func (l *LocalStorage) List(prefix string) ([]string, error) {
	prefix = filepath.Join(l.root, prefix)

	entries, err := os.ReadDir(prefix)
	if err != nil {
		return nil, err
	}

	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		paths = append(paths, filepath.Join(prefix, entry.Name()))
	}

	return paths, nil
}

func (l *LocalStorage) Delete(key string) error {
	path := filepath.Join(l.root, key)

	return os.RemoveAll(path)
}
