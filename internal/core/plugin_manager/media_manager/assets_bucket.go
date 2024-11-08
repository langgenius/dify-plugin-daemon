package media_manager

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path"
	"path/filepath"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/langgenius/dify-plugin-daemon/internal/oss"
)

type MediaBucket struct {
	oss       oss.OSS
	cache     *lru.Cache[string, []byte]
	mediaPath string
}

func NewAssetsBucket(oss oss.OSS, media_path string, cache_size uint16) *MediaBucket {
	// lru.New only raises error when cache_size is a negative number, which is impossible
	cache, _ := lru.New[string, []byte](int(cache_size))

	return &MediaBucket{oss: oss, cache: cache, mediaPath: media_path}
}

// Upload uploads a file to the media manager and returns an identifier
func (m *MediaBucket) Upload(name string, file []byte) (string, error) {
	// calculate checksum
	checksum := sha256.Sum256(append(file, []byte(name)...))

	id := hex.EncodeToString(checksum[:])

	// get file extension
	ext := filepath.Ext(name)

	filename := id + ext

	// store locally
	filePath := path.Join(m.mediaPath, filename)
	err := os.WriteFile(filePath, file, 0o644)
	if err != nil {
		return "", err
	}

	return filename, nil
}

func (m *MediaBucket) Get(id string) ([]byte, error) {
	// check if id is in cache
	data, ok := m.cache.Get(id)
	if ok {
		return data, nil
	}

	// check if id is in storage
	filePath := path.Join(m.mediaPath, id)
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

func (m *MediaBucket) Delete(id string) error {
	// delete from cache
	m.cache.Remove(id)

	// delete from storage
	filePath := path.Join(m.mediaPath, id)
	return os.Remove(filePath)
}
