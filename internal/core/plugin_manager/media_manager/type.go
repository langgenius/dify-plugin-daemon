package media_manager

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/strings"
)

type MediaManager struct {
	storagePath string
	cache       *lru.Cache[string, []byte]
}

func NewMediaManager(storage_path string, cache_size uint16) *MediaManager {
	// mkdir -p storage_path
	if err := os.MkdirAll(storage_path, 0o755); err != nil {
		log.Error("Failed to create storage path: %s", err)
	}

	// lru.New only raises error when cache_size is a negative number, which is impossible
	cache, _ := lru.New[string, []byte](int(cache_size))

	return &MediaManager{storagePath: storage_path, cache: cache}
}

// Upload uploads a file to the media manager and returns an identifier
func (m *MediaManager) Upload(file []byte) (string, error) {
	// calculate checksum
	checksum := sha256.Sum256(append(file, []byte(strings.RandomString(10))...))

	id := hex.EncodeToString(checksum[:])

	// store locally
	filePath := path.Join(m.storagePath, id)
	err := os.WriteFile(filePath, file, 0o644)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (m *MediaManager) Get(id string) ([]byte, error) {
	// check if id is in cache
	data, ok := m.cache.Get(id)
	if ok {
		return data, nil
	}

	// check if id is in storage
	filePath := path.Join(m.storagePath, id)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, err
	}

	// read file
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// store in cache
	m.cache.Add(id, file)

	return file, nil
}

func (m *MediaManager) Delete(id string) error {
	// delete from cache
	m.cache.Remove(id)

	// delete from storage
	filepath := path.Join(m.storagePath, id)
	return os.Remove(filepath)
}
