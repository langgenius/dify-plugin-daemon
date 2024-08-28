package persistence

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
)

type Persistence struct {
	storage PersistenceStorage
}

const (
	CACHE_KEY_PREFIX = "persistence:cache"
)

func (c *Persistence) getCacheKey(tenant_id string, plugin_identity string, key string) string {
	return fmt.Sprintf("%s:%s:%s:%s", CACHE_KEY_PREFIX, tenant_id, plugin_identity, key)
}

func (c *Persistence) Save(tenant_id string, plugin_identity string, key string, data []byte) error {
	if len(key) > 64 {
		return fmt.Errorf("key length must be less than 64 characters")
	}

	return c.storage.Save(tenant_id, plugin_identity, key, data)
}

func (c *Persistence) Load(tenant_id string, plugin_identity string, key string) ([]byte, error) {
	// check if the key exists in cache
	h, err := cache.GetString(c.getCacheKey(tenant_id, plugin_identity, key))
	if err != nil && err != cache.ErrNotFound {
		return nil, err
	}
	if err == nil {
		return hex.DecodeString(h)
	}

	// load from storage
	data, err := c.storage.Load(tenant_id, plugin_identity, key)
	if err != nil {
		return nil, err
	}

	// add to cache
	cache.Store(c.getCacheKey(tenant_id, plugin_identity, key), hex.EncodeToString(data), time.Minute*5)

	return data, nil
}

func (c *Persistence) Delete(tenant_id string, plugin_identity string, key string) error {
	// delete from cache and storage
	err := cache.Del(c.getCacheKey(tenant_id, plugin_identity, key))
	if err != nil {
		return err
	}
	return c.storage.Delete(tenant_id, plugin_identity, key)
}
