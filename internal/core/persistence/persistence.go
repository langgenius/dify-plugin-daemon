package persistence

import (
	"encoding/hex"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache"
	"gorm.io/gorm"
)

type Persistence struct {
	maxStorageSize int64

	storage PersistenceStorage
}

const (
	CACHE_KEY_PREFIX = "persistence:cache"
)

func (c *Persistence) getCacheKey(tenantId string, pluginId string, key string) string {
	return fmt.Sprintf("%s:%s:%s:%s", CACHE_KEY_PREFIX, tenantId, pluginId, key)
}

func (c *Persistence) checkPathTraversal(key string) error {
	key = path.Clean(key)
	if strings.Contains(key, "..") || strings.Contains(key, "//") || strings.Contains(key, "\\") {
		return fmt.Errorf("invalid key: path traversal attempt detected")
	}
	return nil
}

func (c *Persistence) Save(tenantId string, pluginId string, maxSize int64, key string, data []byte) error {
	if err := c.checkPathTraversal(key); err != nil {
		return err
	}

	if len(key) > 256 {
		return fmt.Errorf("key length must be less than 256 characters")
	}

	if maxSize == -1 {
		maxSize = c.maxStorageSize
	}

	lockKey := fmt.Sprintf("persistence:lock:%s:%s:%s", tenantId, pluginId, key)
	if err := cache.Lock(lockKey, 30*time.Second, 3*time.Second); err != nil {
		return err
	}
	defer func() { _ = cache.Unlock(lockKey) }()

	newSize := int64(len(data))
	var oldSize int64 = 0
	if exist, err := c.storage.Exists(tenantId, pluginId, key); err == nil && exist {
		if s, err2 := c.storage.StateSize(tenantId, pluginId, key); err2 == nil {
			oldSize = s
		}
	}
	delta := newSize - oldSize

	txErr := db.WithTransaction(func(tx *gorm.DB) error {
		record, err := db.GetOne[models.TenantStorage](
			db.WithTransactionContext(tx),
			db.Equal("tenant_id", tenantId),
			db.Equal("plugin_id", pluginId),
			db.WLock(),
		)

		if err != nil {
			if !errors.Is(err, db.ErrDatabaseNotFound) {
				return err
			}
			record = models.TenantStorage{TenantID: tenantId, PluginID: pluginId, Size: 0}
			if cerr := db.Create(&record, tx); cerr != nil {
				return cerr
			}
		}

		if delta > 0 {
			if record.Size+delta > c.maxStorageSize || record.Size+delta > maxSize {
				return fmt.Errorf("allocated size is greater than max storage size")
			}
		}

		if err := c.storage.Save(tenantId, pluginId, key, data); err != nil {
			return err
		}

		if delta != 0 {
			if err := db.Run(
				db.WithTransactionContext(tx),
				db.Model(&models.TenantStorage{}),
				db.Equal("tenant_id", tenantId),
				db.Equal("plugin_id", pluginId),
				db.Inc(map[string]int64{"size": delta}),
			); err != nil {
				return err
			}
		}

		return nil
	})
	if txErr != nil {
		return txErr
	}

	if _, err := cache.Del(c.getCacheKey(tenantId, pluginId, key)); errors.Is(err, cache.ErrNotFound) {
		return nil
	} else {
		return err
	}
}

// TODO: raises specific error to avoid confusion
func (c *Persistence) Load(tenantId string, pluginId string, key string) ([]byte, error) {
	if err := c.checkPathTraversal(key); err != nil {
		return nil, err
	}

	// check if the key exists in cache
	h, err := cache.GetString(c.getCacheKey(tenantId, pluginId, key))
	if err != nil && err != cache.ErrNotFound {
		return nil, err
	}
	if err == nil {
		return hex.DecodeString(h)
	}

	// load from storage
	data, err := c.storage.Load(tenantId, pluginId, key)
	if err != nil {
		return nil, err
	}

	// add to cache
	cache.Store(c.getCacheKey(tenantId, pluginId, key), hex.EncodeToString(data), time.Minute*5)

	return data, nil
}

func (c *Persistence) Delete(tenantId string, pluginId string, key string) (int64, error) {
	lockKey := fmt.Sprintf("persistence:lock:%s:%s:%s", tenantId, pluginId, key)
	if err := cache.Lock(lockKey, 30*time.Second, 3*time.Second); err != nil {
		return 0, err
	}
	defer func() { _ = cache.Unlock(lockKey) }()

	// delete from cache and storage
	deletedNum, err := cache.Del(c.getCacheKey(tenantId, pluginId, key))
	if err != nil {
		return 0, err
	}

	// state size
	size, err := c.storage.StateSize(tenantId, pluginId, key)
	if err != nil {
		return 0, err
	}

	if err = c.storage.Delete(tenantId, pluginId, key); err != nil {
		return 0, err
	}

	if err := db.WithTransaction(func(tx *gorm.DB) error {
		_, _ = db.GetOne[models.TenantStorage](
			db.WithTransactionContext(tx),
			db.Equal("tenant_id", tenantId),
			db.Equal("plugin_id", pluginId),
			db.WLock(),
		)
		return db.Run(
			db.WithTransactionContext(tx),
			db.Model(&models.TenantStorage{}),
			db.Equal("tenant_id", tenantId),
			db.Equal("plugin_id", pluginId),
			db.Inc(map[string]int64{"size": -size}),
		)
	}); err != nil {
		return 0, err
	}

	return deletedNum, nil
}

func (c *Persistence) Exist(tenantId string, pluginId string, key string) (int64, error) {
	existNum, err := cache.Exist(c.getCacheKey(tenantId, pluginId, key))
	if err != nil {
		return 0, err
	}
	if existNum > 0 {
		return existNum, nil
	}

	isExist, err := c.storage.Exists(tenantId, pluginId, key)
	if err != nil {
		return 0, err
	}
	if isExist {
		return 1, nil
	}
	return 0, nil
}
