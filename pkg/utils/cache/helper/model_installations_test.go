package helper

import (
	"strings"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	TEST_TENANT_ID = "8f2c1b6e-3a4d-4f5e-9a7b-1c2d3e4f5a6b"
	// deliberately all different, so confusing any two of them fails a test
	TEST_PLUGIN_UNIQUE_IDENTIFIER = "acme/vertex-connector:0.4.1@0123456789abcdef"
	TEST_PLUGIN_ID                = "acme/vertex-connector"
	TEST_ROW_PROVIDER             = "row-provider"
	TEST_DECLARATION_PROVIDER     = "declaration-provider"
)

func setupModelInstallationsTest(t *testing.T) {
	t.Helper()

	redisServer := miniredis.RunT(t)
	require.NoError(t, cache.InitRedisClient(redisServer.Addr(), cache.RedisCredentials{}, false, 0, nil))
	t.Cleanup(func() {
		_ = cache.Close()
	})

	gormDB, err := gorm.Open(
		sqlite.Open("file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=shared"),
		&gorm.Config{},
	)
	require.NoError(t, err)
	require.NoError(t, gormDB.AutoMigrate(
		&models.AIModelInstallation{},
		&models.PluginDeclaration{},
	))
	db.DifyPluginDB = gormDB
	t.Cleanup(func() {
		if sqlDB, err := gormDB.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})

	require.NoError(t, gormDB.Create(&models.PluginDeclaration{
		PluginUniqueIdentifier: TEST_PLUGIN_UNIQUE_IDENTIFIER,
		PluginID:               TEST_PLUGIN_ID,
		Declaration: plugin_entities.PluginDeclaration{
			Model: &plugin_entities.ModelProviderDeclaration{
				Provider: TEST_DECLARATION_PROVIDER,
			},
		},
	}).Error)

	require.NoError(t, gormDB.Create(&models.AIModelInstallation{
		PluginID:               TEST_PLUGIN_ID,
		PluginUniqueIdentifier: TEST_PLUGIN_UNIQUE_IDENTIFIER,
		TenantID:               TEST_TENANT_ID,
		Provider:               TEST_ROW_PROVIDER,
	}).Error)

	t.Cleanup(func() {
		DeletePluginDeclarationCache(
			plugin_entities.PluginUniqueIdentifier(TEST_PLUGIN_UNIQUE_IDENTIFIER),
			plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL,
		)
	})
}

func deleteInstallationRows(t *testing.T) {
	t.Helper()
	require.NoError(t, db.DifyPluginDB.
		Where("tenant_id = ?", TEST_TENANT_ID).
		Delete(&models.AIModelInstallation{}).Error)
}

func TestCombinedListModelInstallationsServesCachedPageAfterRowsChange(t *testing.T) {
	setupModelInstallationsTest(t)

	first, err := CombinedListModelInstallations(TEST_TENANT_ID, 1, 256)
	require.NoError(t, err)
	require.Len(t, first, 1)
	assert.Equal(t, TEST_ROW_PROVIDER, first[0].Provider)
	assert.Equal(t, TEST_PLUGIN_ID, first[0].PluginID)
	require.NotNil(t, first[0].Declaration)
	assert.Equal(t, TEST_DECLARATION_PROVIDER, first[0].Declaration.Provider)

	deleteInstallationRows(t)

	cached, err := CombinedListModelInstallations(TEST_TENANT_ID, 1, 256)
	require.NoError(t, err)
	assert.Len(t, cached, 1, "a cache hit must not reach the database")

	DeleteModelInstallationsCache(TEST_TENANT_ID)

	fresh, err := CombinedListModelInstallations(TEST_TENANT_ID, 1, 256)
	require.NoError(t, err)
	assert.Empty(t, fresh)
}

func TestCombinedListModelInstallationsCachesEachPageSeparately(t *testing.T) {
	setupModelInstallationsTest(t)

	firstPage, err := CombinedListModelInstallations(TEST_TENANT_ID, 1, 256)
	require.NoError(t, err)
	require.Len(t, firstPage, 1)

	secondPage, err := CombinedListModelInstallations(TEST_TENANT_ID, 2, 256)
	require.NoError(t, err)
	assert.Empty(t, secondPage, "page 2 must not be served the page 1 entry")

	DeleteModelInstallationsCache(TEST_TENANT_ID)
	deleteInstallationRows(t)

	afterInvalidation, err := CombinedListModelInstallations(TEST_TENANT_ID, 1, 256)
	require.NoError(t, err)
	assert.Empty(t, afterInvalidation, "one delete must drop every cached page")
}
