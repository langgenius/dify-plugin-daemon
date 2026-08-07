package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/manifest_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/stretchr/testify/require"
)

func TestListPluginsByCategoryRequestBindsFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	var received listPluginsByCategoryRequest
	router.GET("/plugin/:tenant_id/:category/list", func(c *gin.Context) {
		BindRequest(c, func(request listPluginsByCategoryRequest) {
			received = request
			c.Status(http.StatusNoContent)
		})
	})

	request := httptest.NewRequest(
		http.MethodGet,
		"/plugin/tenant-1/tool/list?page=2&page_size=25&query=Weather&tags=search&tags=rag&language=zh_Hans",
		nil,
	)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusNoContent, response.Code)
	require.Equal(t, "tenant-1", received.TenantID)
	require.Equal(t, plugin_entities.PLUGIN_CATEGORY_TOOL, received.Category)
	require.Equal(t, 2, received.Page)
	require.Equal(t, 25, received.PageSize)
	require.Equal(t, "Weather", received.Query)
	require.Equal(
		t,
		[]manifest_entities.PluginTag{
			manifest_entities.PLUGIN_TAG_SEARCH,
			manifest_entities.PLUGIN_TAG_RAG,
		},
		received.Tags,
	)
	require.Equal(t, "zh_Hans", received.Language)
}

func TestListInstalledPluginIDsRequestBindsCategory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	var received listInstalledPluginIDsRequest
	router.GET("/plugin/:tenant_id/installation/ids", func(c *gin.Context) {
		BindRequest(c, func(request listInstalledPluginIDsRequest) {
			received = request
			c.Status(http.StatusNoContent)
		})
	})

	request := httptest.NewRequest(
		http.MethodGet,
		"/plugin/tenant-1/installation/ids?category=tool",
		nil,
	)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusNoContent, response.Code)
	require.Equal(t, "tenant-1", received.TenantID)
	require.Equal(t, plugin_entities.PLUGIN_CATEGORY_TOOL, received.Category)
}
