package media_transport

import (
	"crypto/sha256"
	"encoding/hex"
	"path"
	"path/filepath"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/langgenius/dify-cloud-kit/oss"
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
	err := m.oss.Save(filePath, file)
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
	file, err := m.oss.Load(filePath)
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
	return m.oss.Delete(filePath)
}
