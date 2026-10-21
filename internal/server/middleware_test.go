package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/server/constants"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models/curd"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestFetchPluginInstallationBindsRequestedIdentifier(t *testing.T) {
	gin.SetMode(gin.TestMode)
	redisServer := miniredis.RunT(t)
	require.NoError(t, cache.InitRedisClient(redisServer.Addr(), cache.RedisCredentials{}, false, 0, nil))
	t.Cleanup(func() { require.NoError(t, cache.Close()) })
	installationDB, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, installationDB.AutoMigrate(&models.PluginInstallation{}, &models.Plugin{}, &models.PluginDeclaration{}))
	originalDB := db.DifyPluginDB
	db.DifyPluginDB = installationDB
	t.Cleanup(func() {
		db.DifyPluginDB = originalDB
		sqlDB, err := installationDB.DB()
		require.NoError(t, err)
		require.NoError(t, sqlDB.Close())
	})

	const pluginID = "langgenius/test"
	const oldIdentifier = pluginID + ":1.0.0@aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	const currentIdentifier = pluginID + ":2.0.0@bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	installation := models.PluginInstallation{
		TenantID: "tenant-a", PluginID: pluginID, PluginUniqueIdentifier: currentIdentifier,
		RuntimeType: string(plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL),
	}
	require.NoError(t, installationDB.Create(&installation).Error)
	require.NoError(t, cache.AutoSet(helper.PluginInstallationCacheKey(pluginID, "tenant-a"), installation))
	otherInstallation := models.PluginInstallation{
		TenantID: "tenant-b", PluginID: pluginID, PluginUniqueIdentifier: oldIdentifier,
		RuntimeType: string(plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE),
	}
	require.NoError(t, installationDB.Create(&otherInstallation).Error)
	require.NoError(t, cache.AutoSet(helper.PluginInstallationCacheKey(pluginID, "tenant-b"), otherInstallation))
	require.NoError(t, installationDB.Create(&models.Plugin{
		PluginID: pluginID, PluginUniqueIdentifier: oldIdentifier, Refers: 1,
		InstallType: plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE,
	}).Error)
	databaseQueries := 0
	require.NoError(t, installationDB.Callback().Query().Before("gorm:query").Register("count_installation_queries", func(_ *gorm.DB) {
		databaseQueries++
	}))

	for _, tt := range []struct {
		name       string
		pluginID   string
		identifier string
		tenantID   string
		status     int
	}{
		{"installed version", pluginID, currentIdentifier, "tenant-a", http.StatusOK},
		{"prepared version differs from installed version", pluginID, oldIdentifier, "tenant-a", http.StatusOK},
		{"different installed version", pluginID, oldIdentifier, "tenant-b", http.StatusOK},
		{"requested checksum is preserved", pluginID, pluginID + ":2.0.0@cccccccccccccccccccccccccccccccc", "tenant-a", http.StatusOK},
		{"plugin installed only in another tenant", pluginID, oldIdentifier, "tenant-c", http.StatusNotFound},
		{"plugin not installed", "langgenius/other", "langgenius/other:1.0.0@aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "tenant-a", http.StatusNotFound},
		{"legacy client without identifier", pluginID, "", "tenant-a", http.StatusOK},
		{"malformed identifier", pluginID, "invalid", "tenant-a", http.StatusBadRequest},
		{"different plugin", "langgenius/other", oldIdentifier, "tenant-a", http.StatusBadRequest},
		{"missing plugin id", "", oldIdentifier, "tenant-a", http.StatusBadRequest},
	} {
		t.Run(tt.name, func(t *testing.T) {
			app := &App{}
			router := gin.New()
			dispatched := false
			router.POST("/plugin/:tenant_id/dispatch/tool/invoke", app.FetchPluginInstallation(), func(ctx *gin.Context) {
				dispatched = true
				expectedIdentifier := tt.identifier
				if expectedIdentifier == "" {
					expectedIdentifier = currentIdentifier
				}
				assert.Equal(t, plugin_entities.PluginUniqueIdentifier(expectedIdentifier), ctx.MustGet(constants.CONTEXT_KEY_PLUGIN_UNIQUE_IDENTIFIER))
				expectedInstallation := installation
				if tt.tenantID == "tenant-b" {
					expectedInstallation = otherInstallation
				}
				actualInstallation := ctx.MustGet(constants.CONTEXT_KEY_PLUGIN_INSTALLATION).(models.PluginInstallation)
				assert.Equal(t, expectedInstallation.ID, actualInstallation.ID)
				assert.Equal(t, expectedInstallation.TenantID, actualInstallation.TenantID)
				assert.Equal(t, expectedInstallation.PluginID, actualInstallation.PluginID)
				assert.Equal(t, expectedInstallation.PluginUniqueIdentifier, actualInstallation.PluginUniqueIdentifier)
				assert.Equal(t, expectedInstallation.RuntimeType, actualInstallation.RuntimeType)
				ctx.Status(http.StatusOK)
			})
			req := httptest.NewRequest(http.MethodPost, "/plugin/"+tt.tenantID+"/dispatch/tool/invoke", nil)
			req.Header.Set(constants.X_PLUGIN_ID, tt.pluginID)
			if tt.identifier != "" {
				req.Header.Set(constants.PluginUniqueIdentifier, tt.identifier)
			}
			response := httptest.NewRecorder()
			queriesBefore := databaseQueries
			router.ServeHTTP(response, req)
			assert.Equal(t, tt.status, response.Code)
			assert.Equal(t, tt.status == http.StatusOK, dispatched)
			expectedQueries := 0
			if tt.status == http.StatusNotFound {
				expectedQueries = 1
			}
			assert.Equal(t, expectedQueries, databaseQueries-queriesBefore)
		})
	}

	t.Run("pinned requests populate the cache shared with legacy calls", func(t *testing.T) {
		cacheKey := helper.PluginInstallationCacheKey(pluginID, "tenant-a")
		_, err := cache.AutoDelete[models.PluginInstallation](cacheKey)
		require.NoError(t, err)
		app := &App{}
		router := gin.New()
		var expectedIdentifier string
		router.POST("/plugin/:tenant_id/dispatch/tool/invoke", app.FetchPluginInstallation(), func(ctx *gin.Context) {
			assert.Equal(t, plugin_entities.PluginUniqueIdentifier(expectedIdentifier), ctx.MustGet(constants.CONTEXT_KEY_PLUGIN_UNIQUE_IDENTIFIER))
			ctx.Status(http.StatusOK)
		})
		for index, identifier := range []string{oldIdentifier, "", pluginID + ":2.0.0@cccccccccccccccccccccccccccccccc"} {
			expectedIdentifier = identifier
			if identifier == "" {
				expectedIdentifier = currentIdentifier
			}
			req := httptest.NewRequest(http.MethodPost, "/plugin/tenant-a/dispatch/tool/invoke", nil)
			req.Header.Set(constants.X_PLUGIN_ID, pluginID)
			req.Header.Set(constants.PluginUniqueIdentifier, identifier)
			response := httptest.NewRecorder()
			queriesBefore := databaseQueries
			router.ServeHTTP(response, req)
			assert.Equal(t, http.StatusOK, response.Code)
			expectedQueries := 0
			if index == 0 {
				expectedQueries = 1
			}
			assert.Equal(t, expectedQueries, databaseQueries-queriesBefore)
		}
		cached, err := cache.AutoGet[models.PluginInstallation](cacheKey)
		require.NoError(t, err)
		assert.Equal(t, currentIdentifier, cached.PluginUniqueIdentifier)
	})

	t.Run("upgrade preserves the prepared identifier despite a stale installation cache", func(t *testing.T) {
		require.NoError(t, installationDB.Model(&installation).Update("plugin_unique_identifier", oldIdentifier).Error)
		cachedInstallation := installation
		cachedInstallation.PluginUniqueIdentifier = oldIdentifier
		require.NoError(t, cache.AutoSet(helper.PluginInstallationCacheKey(pluginID, "tenant-a"), cachedInstallation))
		require.NoError(t, installationDB.Model(&installation).Update("plugin_unique_identifier", currentIdentifier).Error)

		app := &App{}
		router := gin.New()
		dispatched := false
		expectedInstallationIdentifier := oldIdentifier
		router.POST("/plugin/:tenant_id/dispatch/tool/invoke", app.FetchPluginInstallation(), func(ctx *gin.Context) {
			dispatched = true
			assert.Equal(t, plugin_entities.PluginUniqueIdentifier(oldIdentifier), ctx.MustGet(constants.CONTEXT_KEY_PLUGIN_UNIQUE_IDENTIFIER))
			actualInstallation := ctx.MustGet(constants.CONTEXT_KEY_PLUGIN_INSTALLATION).(models.PluginInstallation)
			assert.Equal(t, expectedInstallationIdentifier, actualInstallation.PluginUniqueIdentifier)
			ctx.Status(http.StatusOK)
		})
		req := httptest.NewRequest(http.MethodPost, "/plugin/tenant-a/dispatch/tool/invoke", nil)
		req.Header.Set(constants.X_PLUGIN_ID, pluginID)
		req.Header.Set(constants.PluginUniqueIdentifier, oldIdentifier)
		response := httptest.NewRecorder()
		queriesBefore := databaseQueries
		router.ServeHTTP(response, req)
		assert.Equal(t, http.StatusOK, response.Code)
		assert.True(t, dispatched)
		assert.Equal(t, 0, databaseQueries-queriesBefore)

		_, err := cache.AutoDelete[models.PluginInstallation](helper.PluginInstallationCacheKey(pluginID, "tenant-a"))
		require.NoError(t, err)
		expectedInstallationIdentifier = currentIdentifier
		dispatched = false
		response = httptest.NewRecorder()
		router.ServeHTTP(response, req)
		assert.Equal(t, http.StatusOK, response.Code)
		assert.True(t, dispatched)
		assert.Equal(t, 1, databaseQueries-queriesBefore)
	})

	t.Run("uninstalled version is not authorized by a previous lookup", func(t *testing.T) {
		_, err := curd.UninstallPlugin("tenant-b", plugin_entities.PluginUniqueIdentifier(oldIdentifier), otherInstallation.ID, &plugin_entities.PluginDeclaration{})
		require.NoError(t, err)
		_, err = cache.AutoGet[models.PluginInstallation](helper.PluginInstallationCacheKey(pluginID, "tenant-b"))
		require.Error(t, err)
		app := &App{}
		router := gin.New()
		router.POST("/plugin/:tenant_id/dispatch/tool/invoke", app.FetchPluginInstallation(), func(ctx *gin.Context) {
			t.Error("Uninstalled plugin must not be dispatched")
		})
		req := httptest.NewRequest(http.MethodPost, "/plugin/tenant-b/dispatch/tool/invoke", nil)
		req.Header.Set(constants.X_PLUGIN_ID, pluginID)
		req.Header.Set(constants.PluginUniqueIdentifier, oldIdentifier)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		assert.Equal(t, http.StatusNotFound, response.Code)
	})
}

func TestPrometheusMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(PrometheusMiddleware())

	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test")
	})

	router.POST("/api/test", func(c *gin.Context) {
		c.String(http.StatusCreated, "created")
	})

	router.GET("/error", func(c *gin.Context) {
		c.String(http.StatusInternalServerError, "error")
	})

	t.Run("GET request", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "test", w.Body.String())
	})

	t.Run("POST request", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "created", w.Body.String())
	})

	t.Run("Error request", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/error", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "error", w.Body.String())
	})
}
