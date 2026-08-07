package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestListModelPluginBindingsRequestBindsTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	var received listModelPluginBindingsRequest
	router.GET("/plugin/:tenant_id/management/models/bindings", func(c *gin.Context) {
		BindRequest(c, func(request listModelPluginBindingsRequest) {
			received = request
			c.Status(http.StatusNoContent)
		})
	})

	request := httptest.NewRequest(
		http.MethodGet,
		"/plugin/tenant-1/management/models/bindings",
		nil,
	)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusNoContent, response.Code)
	require.Equal(t, "tenant-1", received.TenantID)
}
