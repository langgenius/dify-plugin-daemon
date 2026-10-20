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
	require.NoError(t, installationDB.AutoMigrate(&models.PluginInstallation{}))
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

	for _, tt := range []struct {
		name       string
		pluginID   string
		identifier string
		tenantID   string
		status     int
	}{
		{"installed version", pluginID, currentIdentifier, "tenant-a", http.StatusOK},
		{"version installed in another tenant", pluginID, oldIdentifier, "tenant-a", http.StatusNotFound},
		{"different installed version", pluginID, oldIdentifier, "tenant-b", http.StatusOK},
		{"different checksum", pluginID, pluginID + ":2.0.0@cccccccccccccccccccccccccccccccc", "tenant-a", http.StatusNotFound},
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
			router.ServeHTTP(response, req)
			assert.Equal(t, tt.status, response.Code)
			assert.Equal(t, tt.status == http.StatusOK, dispatched)
		})
	}

	t.Run("uninstalled version is not authorized by a previous lookup", func(t *testing.T) {
		require.NoError(t, installationDB.Delete(&otherInstallation).Error)
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
