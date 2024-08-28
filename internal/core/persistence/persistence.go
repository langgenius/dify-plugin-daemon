package persistence

import (
	"encoding/hex"
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
)

type Persistence struct {
	storage PersistenceStorage
}

const (
	CACHE_KEY_PREFIX = "persistence:cache"
)

func (c *Persistence) getCacheKey(tenant_id string, plugin_checksum string) string {
	return fmt.Sprintf("%s:%s:%s", CACHE_KEY_PREFIX, tenant_id, plugin_checksum)
}

func (c *Persistence) Save(tenant_id string, plugin_checksum string, key string, data []byte) error {
	// add to cache
	h := hex.EncodeToString(data)
	return cache.SetMapOneField(c.getCacheKey(tenant_id, plugin_checksum), key, h)
}

func (c *Persistence) Load(tenant_id string, plugin_checksum string, key string) ([]byte, error) {
	// check if the key exists in cache
	h, err := cache.GetMapFieldString(c.getCacheKey(tenant_id, plugin_checksum), key)
	if err != nil && err != cache.ErrNotFound {
		return nil, err
	}
	if err == nil {
		return hex.DecodeString(h)
	}

	// load from storage
	return c.storage.Load(tenant_id, plugin_checksum, key)
}

func (c *Persistence) Delete(tenant_id string, plugin_checksum string, key string) error {
	// delete from cache and storage
	err := cache.DelMapField(c.getCacheKey(tenant_id, plugin_checksum), key)
	if err != nil {
		return err
	}
	return c.storage.Delete(tenant_id, plugin_checksum, key)
}
