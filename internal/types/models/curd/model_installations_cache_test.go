package curd

import (
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	CACHE_TEST_TENANT_ID                = "3d9e7c11-5b2a-4f8d-9e01-7a6b5c4d3e2f"
	CACHE_TEST_PLUGIN_UNIQUE_IDENTIFIER = "acme/vertex-connector:0.4.1@0123456789abcdef"
	CACHE_TEST_PLUGIN_ID                = "acme/vertex-connector"
	CACHE_TEST_PAGE_FIELD               = "1:256"
	CACHE_TEST_SENTINEL                 = `[{"provider":"stale-entry"}]`
)

func cleanupInstallPluginCacheTest(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		_ = db.DeleteByCondition(&models.PluginInstallation{TenantID: CACHE_TEST_TENANT_ID})
		_ = db.DeleteByCondition(&models.AIModelInstallation{TenantID: CACHE_TEST_TENANT_ID})
		_ = db.DeleteByCondition(&models.Plugin{
			PluginUniqueIdentifier: CACHE_TEST_PLUGIN_UNIQUE_IDENTIFIER,
		})
		_, _ = cache.Del(helper.ModelInstallationsCacheKey(CACHE_TEST_TENANT_ID))
	})
}

func seedModelInstallationsCache(t *testing.T) {
	t.Helper()
	require.NoError(t, cache.SetMapOneField(
		helper.ModelInstallationsCacheKey(CACHE_TEST_TENANT_ID),
		CACHE_TEST_PAGE_FIELD,
		CACHE_TEST_SENTINEL,
	))
	require.True(t, modelInstallationsCacheExists(t), "seed must land before the assertion is meaningful")
}

func modelInstallationsCacheExists(t *testing.T) bool {
	t.Helper()
	exists, err := cache.Exist(helper.ModelInstallationsCacheKey(CACHE_TEST_TENANT_ID))
	require.NoError(t, err)
	return exists == 1
}

func installCacheTestPlugin(declaration *plugin_entities.PluginDeclaration) error {
	_, _, err := InstallPlugin(
		CACHE_TEST_TENANT_ID,
		plugin_entities.PluginUniqueIdentifier(CACHE_TEST_PLUGIN_UNIQUE_IDENTIFIER),
		plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL,
		declaration,
		"test",
		nil,
	)
	return err
}

// The invalidation must also run when the transaction fails, otherwise a retry after a
// failed Redis DEL can never clear the cache: it aborts on ErrPluginAlreadyInstalled first.
func TestInstallPluginInvalidatesModelInstallationsCacheOnRetry(t *testing.T) {
	cleanupInstallPluginCacheTest(t)

	declaration := &plugin_entities.PluginDeclaration{
		Model: &plugin_entities.ModelProviderDeclaration{
			Provider: "declaration-provider",
		},
	}

	seedModelInstallationsCache(t)
	require.NoError(t, installCacheTestPlugin(declaration))
	assert.False(t, modelInstallationsCacheExists(t), "a successful install must drop the tenant cache")

	seedModelInstallationsCache(t)
	require.ErrorIs(t, installCacheTestPlugin(declaration), ErrPluginAlreadyInstalled)
	assert.False(t, modelInstallationsCacheExists(t), "a retry must still drop the tenant cache")
}

func TestInstallPluginLeavesModelInstallationsCacheAloneWithoutModel(t *testing.T) {
	cleanupInstallPluginCacheTest(t)

	seedModelInstallationsCache(t)
	require.NoError(t, installCacheTestPlugin(&plugin_entities.PluginDeclaration{}))
	assert.True(t, modelInstallationsCacheExists(t), "a plugin without a model must not invalidate")
}
