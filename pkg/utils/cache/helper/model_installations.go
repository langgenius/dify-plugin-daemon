package helper

import (
	"errors"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

const DEFAULT_MODEL_INSTALLATIONS_CACHE_TTL = 60 * time.Minute

var modelInstallationsCacheTTL atomic.Int64

func SetModelInstallationsCacheTTL(ttl time.Duration) {
	if ttl <= 0 {
		ttl = DEFAULT_MODEL_INSTALLATIONS_CACHE_TTL
	}
	modelInstallationsCacheTTL.Store(int64(ttl))
}

func currentModelInstallationsCacheTTL() time.Duration {
	if ttl := modelInstallationsCacheTTL.Load(); ttl > 0 {
		return time.Duration(ttl)
	}
	return DEFAULT_MODEL_INSTALLATIONS_CACHE_TTL
}

type AIModelInstallationWithDeclaration struct {
	models.AIModelInstallation

	Declaration *plugin_entities.ModelProviderDeclaration `json:"declaration"`
}

func modelInstallationsCacheField(page int, pageSize int) string {
	return strings.Join(
		[]string{
			strconv.Itoa(page),
			strconv.Itoa(pageSize),
		},
		":",
	)
}

func CombinedListModelInstallations(
	tenantId string,
	page int,
	pageSize int,
) ([]AIModelInstallationWithDeclaration, error) {
	cacheKey := ModelInstallationsCacheKey(tenantId)
	cacheField := modelInstallationsCacheField(page, pageSize)

	cached, err := cache.GetMapField[[]AIModelInstallationWithDeclaration](cacheKey, cacheField)
	if err == nil {
		return *cached, nil
	}
	if !errors.Is(err, cache.ErrNotFound) {
		log.Warn("failed to read model installations cache", "key", cacheKey, "error", err)
	}

	installations, err := db.GetAll[models.AIModelInstallation](
		db.Equal("tenant_id", tenantId),
		db.Page(page, pageSize),
	)
	if err != nil {
		return nil, err
	}

	data := make([]AIModelInstallationWithDeclaration, 0, len(installations))

	for _, installation := range installations {
		uniqueIdentifier := plugin_entities.PluginUniqueIdentifier(installation.PluginUniqueIdentifier)
		runtimeType := plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL
		if uniqueIdentifier.RemoteLike() {
			runtimeType = plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE
		}

		declaration, err := CombinedGetPluginDeclaration(uniqueIdentifier, runtimeType)
		if err != nil {
			return nil, err
		}

		data = append(data, AIModelInstallationWithDeclaration{
			AIModelInstallation: installation,
			Declaration:         declaration.Model,
		})
	}

	if err := cache.SetMapOneField(cacheKey, cacheField, data); err != nil {
		log.Warn("failed to store model installations cache", "key", cacheKey, "error", err)
	} else if _, err := cache.Expire(cacheKey, currentModelInstallationsCacheTTL()); err != nil {
		log.Warn("failed to set model installations cache expiry", "key", cacheKey, "error", err)
	}

	return data, nil
}

func DeleteModelInstallationsCache(tenantId string) {
	cacheKey := ModelInstallationsCacheKey(tenantId)
	if _, err := cache.Del(cacheKey); err != nil {
		log.Warn("failed to clear model installations cache", "key", cacheKey, "error", err)
	}
}
