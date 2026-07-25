package service

import (
	"fmt"
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

func TestListInstalledPluginIDsReturnsCompleteCategoryScopedList(t *testing.T) {
	setupPluginCategoryListTestDB(t)

	baseTime := time.Date(2026, time.July, 23, 12, 0, 0, 0, time.UTC)
	for index := 0; index < 300; index++ {
		pluginID := fmt.Sprintf("author/tool-plugin-%03d", index)
		require.NoError(t, db.Create(&models.ToolInstallation{
			Model:                  models.Model{CreatedAt: baseTime.Add(time.Duration(index) * time.Second)},
			TenantID:               "tenant-1",
			Provider:               fmt.Sprintf("provider-%03d", index),
			PluginID:               pluginID,
			PluginUniqueIdentifier: pluginID + ":1.0.0@" + fmt.Sprintf("%032x", index+1),
		}))
	}
	require.NoError(t, db.Create(&models.ToolInstallation{
		Model:                  models.Model{CreatedAt: baseTime.Add(300 * time.Second)},
		TenantID:               "tenant-1",
		Provider:               "duplicate-provider",
		PluginID:               "author/tool-plugin-299",
		PluginUniqueIdentifier: "author/tool-plugin-299:1.0.0@" + fmt.Sprintf("%032x", 300),
	}))
	require.NoError(t, db.Create(&models.ToolInstallation{
		Model:                  models.Model{CreatedAt: baseTime.Add(301 * time.Second)},
		TenantID:               "tenant-2",
		Provider:               "other-tenant-provider",
		PluginID:               "author/other-tenant-plugin",
		PluginUniqueIdentifier: "author/other-tenant-plugin:1.0.0@" + fmt.Sprintf("%032x", 301),
	}))

	response := ListInstalledPluginIDs("tenant-1", plugin_entities.PLUGIN_CATEGORY_TOOL)

	require.Equal(t, 0, response.Code)
	data := response.Data.(installedPluginIDsResponse)
	require.Len(t, data.PluginIDs, 300)
	require.Equal(t, "author/tool-plugin-299", data.PluginIDs[0])
	require.Equal(t, "author/tool-plugin-000", data.PluginIDs[299])
	require.NotContains(t, data.PluginIDs, "author/other-tenant-plugin")
}

func TestListInstalledPluginIDsUsesCategoryInstallationTables(t *testing.T) {
	setupPluginCategoryListTestDB(t)

	installations := []any{
		&models.AIModelInstallation{
			TenantID: "tenant-1",
			Provider: "model-provider",
			PluginID: "author/model-plugin",
		},
		&models.DatasourceInstallation{
			TenantID: "tenant-1",
			Provider: "datasource-provider",
			PluginID: "author/datasource-plugin",
		},
		&models.AgentStrategyInstallation{
			TenantID: "tenant-1",
			Provider: "agent-strategy-provider",
			PluginID: "author/agent-strategy-plugin",
		},
		&models.TriggerInstallation{
			TenantID: "tenant-1",
			Provider: "trigger-provider",
			PluginID: "author/trigger-plugin",
		},
	}
	for _, installation := range installations {
		require.NoError(t, db.Create(installation))
	}

	tests := []struct {
		name     string
		category plugin_entities.PluginCategory
		pluginID string
	}{
		{name: "model", category: plugin_entities.PLUGIN_CATEGORY_MODEL, pluginID: "author/model-plugin"},
		{
			name:     "datasource",
			category: plugin_entities.PLUGIN_CATEGORY_DATASOURCE,
			pluginID: "author/datasource-plugin",
		},
		{
			name:     "agent strategy",
			category: plugin_entities.PLUGIN_CATEGORY_AGENT_STRATEGY,
			pluginID: "author/agent-strategy-plugin",
		},
		{name: "trigger", category: plugin_entities.PLUGIN_CATEGORY_TRIGGER, pluginID: "author/trigger-plugin"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := ListInstalledPluginIDs("tenant-1", test.category)

			require.Equal(t, 0, response.Code)
			require.Equal(t, []string{test.pluginID}, response.Data.(installedPluginIDsResponse).PluginIDs)
		})
	}
}

func TestListInstalledPluginIDsUsesPluginInstallationsForExtensions(t *testing.T) {
	setupPluginCategoryListTestDB(t)

	baseTime := time.Date(2026, time.July, 23, 12, 0, 0, 0, time.UTC)
	createInstalledPlugin(t, "extension-plugin", "a", baseTime.Add(time.Minute), nil)
	createInstalledPlugin(
		t,
		"tool-plugin",
		"b",
		baseTime,
		[]string{"provider/tool.yaml"},
	)

	response := ListInstalledPluginIDs("tenant-1", plugin_entities.PLUGIN_CATEGORY_EXTENSION)

	require.Equal(t, 0, response.Code)
	require.Equal(
		t,
		[]string{"author/extension-plugin"},
		response.Data.(installedPluginIDsResponse).PluginIDs,
	)
}

func TestListModelPluginBindingsUsesCanonicalInstallationWithoutDeclarationHydration(t *testing.T) {
	setupPluginCategoryListTestDB(t)

	baseTime := time.Date(2026, time.July, 24, 12, 0, 0, 0, time.UTC)
	marketplaceIdentifier := "langgenius/openai:1.2.3@" + strings.Repeat("a", 32)
	remoteIdentifier := "debugger/custom-model:0.9.0@" + strings.Repeat("b", 32)

	marketplaceInstallation := &models.PluginInstallation{
		Model:                  models.Model{CreatedAt: baseTime},
		TenantID:               "tenant-1",
		PluginID:               "langgenius/openai",
		PluginUniqueIdentifier: marketplaceIdentifier,
		RuntimeType:            string(plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL),
		Source:                 "marketplace",
		Meta:                   map[string]any{},
	}
	require.NoError(t, db.Create(marketplaceInstallation))
	require.NoError(t, db.Create(&models.AIModelInstallation{
		TenantID:               "tenant-1",
		Provider:               "openai",
		PluginID:               "langgenius/openai",
		PluginUniqueIdentifier: marketplaceIdentifier,
	}))
	// Historical duplicate projections must not produce duplicate provider bindings.
	require.NoError(t, db.Create(&models.AIModelInstallation{
		TenantID:               "tenant-1",
		Provider:               "openai",
		PluginID:               "langgenius/openai",
		PluginUniqueIdentifier: marketplaceIdentifier,
	}))

	remoteInstallation := &models.PluginInstallation{
		Model:                  models.Model{CreatedAt: baseTime.Add(time.Minute)},
		TenantID:               "tenant-1",
		PluginID:               "debugger/custom-model",
		PluginUniqueIdentifier: remoteIdentifier,
		RuntimeType:            string(plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE),
		Source:                 "remote",
		Meta:                   map[string]any{},
	}
	require.NoError(t, db.Create(remoteInstallation))
	require.NoError(t, db.Create(&models.AIModelInstallation{
		TenantID:               "tenant-1",
		Provider:               "custom-model",
		PluginID:               "debugger/custom-model",
		PluginUniqueIdentifier: remoteIdentifier,
	}))

	// A stale projection from an older version must not bind to the current plugin installation.
	require.NoError(t, db.Create(&models.AIModelInstallation{
		TenantID:               "tenant-1",
		Provider:               "stale-openai",
		PluginID:               "langgenius/openai",
		PluginUniqueIdentifier: "langgenius/openai:1.0.0@" + strings.Repeat("c", 32),
	}))

	// Installations without a model projection and model projections from another tenant are excluded.
	require.NoError(t, db.Create(&models.PluginInstallation{
		Model:                  models.Model{CreatedAt: baseTime.Add(2 * time.Minute)},
		TenantID:               "tenant-1",
		PluginID:               "langgenius/tool-only",
		PluginUniqueIdentifier: "langgenius/tool-only:1.0.0@" + strings.Repeat("d", 32),
		RuntimeType:            string(plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL),
		Source:                 "marketplace",
		Meta:                   map[string]any{},
	}))
	otherTenantIdentifier := "langgenius/other-model:1.0.0@" + strings.Repeat("e", 32)
	require.NoError(t, db.Create(&models.PluginInstallation{
		TenantID:               "tenant-2",
		PluginID:               "langgenius/other-model",
		PluginUniqueIdentifier: otherTenantIdentifier,
		RuntimeType:            string(plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL),
		Source:                 "marketplace",
		Meta:                   map[string]any{},
	}))
	require.NoError(t, db.Create(&models.AIModelInstallation{
		TenantID:               "tenant-2",
		Provider:               "other-model",
		PluginID:               "langgenius/other-model",
		PluginUniqueIdentifier: otherTenantIdentifier,
	}))

	response := ListModelPluginBindings("tenant-1")

	require.Equal(t, 0, response.Code)
	bindings := response.Data.([]modelPluginBindingResponse)
	require.Len(t, bindings, 2)
	require.Equal(t, []modelPluginBindingResponse{
		{
			Provider:               "custom-model",
			InstallationID:         remoteInstallation.ID,
			PluginID:               "debugger/custom-model",
			PluginUniqueIdentifier: remoteIdentifier,
			RuntimeType:            plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE,
			Source:                 "remote",
			Version:                manifest_entities.Version("0.9.0"),
		},
		{
			Provider:               "openai",
			InstallationID:         marketplaceInstallation.ID,
			PluginID:               "langgenius/openai",
			PluginUniqueIdentifier: marketplaceIdentifier,
			RuntimeType:            plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL,
			Source:                 "marketplace",
			Version:                manifest_entities.Version("1.2.3"),
		},
	}, bindings)
}

func TestListModelPluginBindingsReturnsEmptyList(t *testing.T) {
	setupPluginCategoryListTestDB(t)

	response := ListModelPluginBindings("tenant-1")

	require.Equal(t, 0, response.Code)
	bindings := response.Data.([]modelPluginBindingResponse)
	require.NotNil(t, bindings)
	require.Empty(t, bindings)
}

func TestListModelPluginBindingsRejectsMalformedCurrentIdentifier(t *testing.T) {
	setupPluginCategoryListTestDB(t)

	const malformedIdentifier = "langgenius/openai:not-valid"
	require.NoError(t, db.Create(&models.PluginInstallation{
		TenantID:               "tenant-1",
		PluginID:               "langgenius/openai",
		PluginUniqueIdentifier: malformedIdentifier,
		RuntimeType:            string(plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL),
		Source:                 "marketplace",
		Meta:                   map[string]any{},
	}))
	require.NoError(t, db.Create(&models.AIModelInstallation{
		TenantID:               "tenant-1",
		Provider:               "openai",
		PluginID:               "langgenius/openai",
		PluginUniqueIdentifier: malformedIdentifier,
	}))

	response := ListModelPluginBindings("tenant-1")

	require.Equal(t, -400, response.Code)
}

func createInstalledPlugin(
	t *testing.T,
	name string,
	checksumCharacter string,
	createdAt time.Time,
	tools []string,
) {
	t.Helper()

	pluginID := "author/" + name
	identifier := pluginID + ":1.0.0@" + strings.Repeat(checksumCharacter, 32)
	require.NoError(t, db.Create(&models.PluginDeclaration{
		PluginUniqueIdentifier: identifier,
		PluginID:               pluginID,
		Declaration: plugin_entities.PluginDeclaration{
			PluginDeclarationWithoutAdvancedFields: plugin_entities.PluginDeclarationWithoutAdvancedFields{
				Name:        name,
				Label:       plugin_entities.I18nObject{EnUS: name},
				Description: plugin_entities.I18nObject{EnUS: name + " description"},
				Plugins:     plugin_entities.PluginExtensions{Tools: tools},
			},
		},
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
	require.NoError(t, gormDB.AutoMigrate(
		&models.PluginDeclaration{},
		&models.PluginInstallation{},
		&models.ToolInstallation{},
		&models.AIModelInstallation{},
		&models.DatasourceInstallation{},
		&models.AgentStrategyInstallation{},
		&models.TriggerInstallation{},
	))
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
