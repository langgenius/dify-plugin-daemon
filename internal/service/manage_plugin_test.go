package service

import (
	"strings"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/manifest_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestListPluginsByCategoryFiltersBeforePagination(t *testing.T) {
	setupPluginCategoryListTestDB(t)

	baseTime := time.Date(2026, time.July, 23, 12, 0, 0, 0, time.UTC)
	createCategoryListPlugin(t, "extension_match", "a", baseTime.Add(4*time.Minute), nil, nil, "匹配扩展")
	createCategoryListPlugin(
		t,
		"wrong_tag",
		"b",
		baseTime.Add(3*time.Minute),
		[]string{"provider/wrong.yaml"},
		[]manifest_entities.PluginTag{manifest_entities.PLUGIN_TAG_RAG},
		"匹配但标签不符",
	)
	createCategoryListPlugin(
		t,
		"first_match",
		"c",
		baseTime.Add(2*time.Minute),
		[]string{"provider/first.yaml"},
		[]manifest_entities.PluginTag{manifest_entities.PLUGIN_TAG_SEARCH},
		"第一个匹配项",
	)
	createCategoryListPlugin(
		t,
		"second_match",
		"d",
		baseTime.Add(time.Minute),
		[]string{"provider/second.yaml"},
		[]manifest_entities.PluginTag{manifest_entities.PLUGIN_TAG_SEARCH},
		"第二个匹配项",
	)

	firstResponse := ListPluginsByCategory(
		"tenant-1",
		plugin_entities.PLUGIN_CATEGORY_TOOL,
		1,
		1,
		"匹配",
		[]manifest_entities.PluginTag{
			manifest_entities.PLUGIN_TAG_WEATHER,
			manifest_entities.PLUGIN_TAG_SEARCH,
		},
		"zh_Hans",
	)

	require.Equal(t, 0, firstResponse.Code)
	firstPage := firstResponse.Data.(pluginListResponse)
	require.True(t, firstPage.HasMore)
	require.Len(t, firstPage.List, 1)
	require.Equal(t, "author/first_match", firstPage.List[0].PluginID)

	secondResponse := ListPluginsByCategory(
		"tenant-1",
		plugin_entities.PLUGIN_CATEGORY_TOOL,
		2,
		1,
		"匹配",
		[]manifest_entities.PluginTag{
			manifest_entities.PLUGIN_TAG_WEATHER,
			manifest_entities.PLUGIN_TAG_SEARCH,
		},
		"zh_Hans",
	)

	require.Equal(t, 0, secondResponse.Code)
	secondPage := secondResponse.Data.(pluginListResponse)
	require.False(t, secondPage.HasMore)
	require.Len(t, secondPage.List, 1)
	require.Equal(t, "author/second_match", secondPage.List[0].PluginID)
}

func setupPluginCategoryListTestDB(t *testing.T) {
	t.Helper()

	redisServer := miniredis.RunT(t)
	require.NoError(t, cache.InitRedisClient(redisServer.Addr(), "", "", false, 0, nil))
	t.Cleanup(func() {
		_ = cache.Close()
	})

	gormDB, err := gorm.Open(sqlite.Open("file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, gormDB.AutoMigrate(&models.PluginDeclaration{}, &models.PluginInstallation{}))
	db.DifyPluginDB = gormDB
	t.Cleanup(func() {
		sqlDB, err := gormDB.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})
}

func createCategoryListPlugin(
	t *testing.T,
	name string,
	checksumCharacter string,
	createdAt time.Time,
	tools []string,
	tags []manifest_entities.PluginTag,
	zhHansLabel string,
) {
	t.Helper()

	pluginID := "author/" + name
	identifier := pluginID + ":1.0.0@" + strings.Repeat(checksumCharacter, 32)
	declaration := plugin_entities.PluginDeclaration{
		PluginDeclarationWithoutAdvancedFields: plugin_entities.PluginDeclarationWithoutAdvancedFields{
			Name: name,
			Label: plugin_entities.I18nObject{
				EnUS:   name,
				ZhHans: zhHansLabel,
			},
			Description: plugin_entities.I18nObject{EnUS: name + " description"},
			Plugins:     plugin_entities.PluginExtensions{Tools: tools},
			Tags:        tags,
		},
	}

	require.NoError(t, db.Create(&models.PluginDeclaration{
		PluginUniqueIdentifier: identifier,
		PluginID:               pluginID,
		Declaration:            declaration,
	}))
	require.NoError(t, db.Create(&models.PluginInstallation{
		Model: models.Model{
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		},
		TenantID:               "tenant-1",
		PluginID:               pluginID,
		PluginUniqueIdentifier: identifier,
		RuntimeType:            string(plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL),
		Meta:                   map[string]any{},
	}))
}
