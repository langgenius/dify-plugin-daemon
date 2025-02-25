package helper

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

var (
	ErrPluginNotFound = errors.New("plugin not found")
)

type memCacheItem struct {
	declaration *plugin_entities.PluginDeclaration
	accessCount int64
	lastAccess  time.Time
}

type memCache struct {
	sync.RWMutex
	items    map[string]*memCacheItem
	itemSize int64
}

var (
	// 500MB memory cache
	maxMemCacheSize = int64(1024)
	// 600s TTL
	maxTTL = 600 * time.Second

	pluginCache = &memCache{
		items:    make(map[string]*memCacheItem),
		itemSize: 0,
	}
)

func (c *memCache) get(key string) *plugin_entities.PluginDeclaration {
	c.RLock()
	item, exists := c.items[key]
	c.RUnlock()

	if !exists {
		return nil
	}

	// Check TTL with a read lock first
	if time.Since(item.lastAccess) > maxTTL {
		c.Lock()
		// Double check after acquiring write lock
		if item, exists = c.items[key]; exists {
			if time.Since(item.lastAccess) > maxTTL {
				c.itemSize--
				delete(c.items, key)
			}
		}
		c.Unlock()
		return nil
	}

	// Update access count and time atomically
	c.Lock()
	if item, exists = c.items[key]; exists {
		item.accessCount++
		item.lastAccess = time.Now()
	}
	c.Unlock()

	if exists {
		return item.declaration
	}
	return nil
}

func (c *memCache) set(key string, declaration *plugin_entities.PluginDeclaration) {
	c.Lock()
	defer c.Unlock()

	// Clean expired items first
	now := time.Now()
	for k, v := range c.items {
		if now.Sub(v.lastAccess) > maxTTL {
			c.itemSize--
			delete(c.items, k)
		}
	}

	// Remove least accessed items if cache is full
	for c.itemSize >= maxMemCacheSize {
		var leastKey string
		var leastCount int64 = -1
		var oldestAccess = time.Now()

		for k, v := range c.items {
			// Prioritize by access count, then by age
			if leastCount == -1 || v.accessCount < leastCount ||
				(v.accessCount == leastCount && v.lastAccess.Before(oldestAccess)) {
				leastCount = v.accessCount
				oldestAccess = v.lastAccess
				leastKey = k
			}
		}

		if leastKey != "" {
			c.itemSize--
			delete(c.items, leastKey)
		}
	}

	// Add new item
	c.items[key] = &memCacheItem{
		declaration: declaration,
		accessCount: 1,
		lastAccess:  now,
	}
	c.itemSize++
}

func CombinedGetPluginDeclaration(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	runtimeType plugin_entities.PluginRuntimeType,
) (*plugin_entities.PluginDeclaration, error) {
	cacheKey := strings.Join(
		[]string{
			"declaration_cache",
			string(runtimeType),
			pluginUniqueIdentifier.String(),
		},
		":",
	)

	// Try memory cache first
	if declaration := pluginCache.get(cacheKey); declaration != nil {
		return declaration, nil
	}

	// Try Redis cache next
	declaration, err := cache.AutoGetWithGetter(
		cacheKey,
		func() (*plugin_entities.PluginDeclaration, error) {
			if runtimeType != plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE {
				declaration, err := db.GetOne[models.PluginDeclaration](
					db.Equal("plugin_unique_identifier", pluginUniqueIdentifier.String()),
				)
				if err == db.ErrDatabaseNotFound {
					return nil, ErrPluginNotFound
				}

				if err != nil {
					return nil, err
				}

				return &declaration.Declaration, nil
			} else {
				// try to fetch the declaration from plugin if it's remote
				plugin, err := db.GetOne[models.Plugin](
					db.Equal("plugin_unique_identifier", pluginUniqueIdentifier.String()),
					db.Equal("install_type", string(plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE)),
				)
				if err == db.ErrDatabaseNotFound {
					return nil, ErrPluginNotFound
				}

				if err != nil {
					return nil, err
				}

				return &plugin.RemoteDeclaration, nil
			}
		},
	)

	if err == nil {
		// Store successful result in memory cache
		pluginCache.set(cacheKey, declaration)
	}

	return declaration, err
}
