package local

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/internal/oss"
)

type LocalStorage struct {
	root string
}

func NewLocalStorage(root string) oss.OSS {
	if err := os.MkdirAll(root, 0o755); err != nil {
		log.Panicf("Failed to create storage path: %s", err)
	}

	return &LocalStorage{root: root}
}

func (l *LocalStorage) Save(key string, data []byte) error {
	path := filepath.Join(l.root, key)
	filePath := filepath.Dir(path)
	if err := os.MkdirAll(filePath, 0o755); err != nil {
		return err
	}

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

func (l *LocalStorage) List(prefix string) ([]oss.OSSPath, error) {
	paths := make([]oss.OSSPath, 0)
	// check if the patch exists
	exists, err := l.Exists(prefix)
	if err != nil {
		return nil, err
	}
	if !exists {
		return paths, nil
	}
	prefix = filepath.Join(l.root, prefix)

	err = filepath.WalkDir(prefix, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// remove prefix
		path = strings.TrimPrefix(path, prefix)
		if path == "" {
			return nil
		}
		// remove leading slash
		path = strings.TrimPrefix(path, "/")
		paths = append(paths, oss.OSSPath{
			Path:  path,
			IsDir: d.IsDir(),
		})
		return nil
	})

	if err != nil {
		return nil, err
	}

	return paths, nil
}

func (l *LocalStorage) Delete(key string) error {
	path := filepath.Join(l.root, key)

	return os.RemoveAll(path)
}

func (l *LocalStorage) Type() string {
	return oss.OSS_TYPE_LOCAL
}