package curd

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache/helper"
	"github.com/stretchr/testify/require"
)

// Test that InstallPlugin invalidates PluginInstallation cache key so that
// subsequent reads fall back to DB instead of returning stale data.
func TestInstallPlugin_InvalidateInstallationCache(t *testing.T) {
	tenantID := uuid.NewString()
	pluginName := "cache_invalidate_" + uuid.NewString()
	checksum := strings.ReplaceAll(uuid.NewString(), "-", "")
	if len(checksum) > 32 {
		checksum = checksum[:32]
	}

	identifier, err := plugin_entities.NewPluginUniqueIdentifier("tester/" + pluginName + ":1.0.0.0@" + checksum)
	require.NoError(t, err)
	pluginID := identifier.PluginID()

	// Seed a stale cache value for the installation key
	key := helper.PluginInstallationCacheKey(pluginID, tenantID)
	require.NoError(t, cache.AutoSet[models.PluginInstallation](key, models.PluginInstallation{
		Model:                  models.Model{ID: "OLD"},
		PluginID:               pluginID,
		PluginUniqueIdentifier: identifier.String(),
		TenantID:               tenantID,
		RuntimeType:            string(plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL),
	}))
	// Perform install which should invalidate the cache key
	_, _, err = InstallPlugin(
		tenantID,
		identifier,
		plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL,
		&plugin_entities.PluginDeclaration{},
		"unittest",
		map[string]any{"from": "test"},
	)
	require.NoError(t, err)

	// Read using AutoGetWithGetter — should miss cache, call getter (DB), then set fresh cache
	inst, err := cache.AutoGetWithGetter(key, func() (*models.PluginInstallation, error) {
		v, e := db.GetOne[models.PluginInstallation](
			db.Equal("plugin_id", pluginID),
			db.Equal("tenant_id", tenantID),
		)
		if e != nil {
			return nil, e
		}
		return &v, nil
	})
	require.NoError(t, err)
	require.NotNil(t, inst)
	require.NotEqual(t, "OLD", inst.ID, "cache should have been invalidated and refetched from DB")
}
