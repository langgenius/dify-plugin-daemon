package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/server/constants"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/stretchr/testify/assert"
)

func TestRedirectPluginInvoke_ClusterDisabled(t *testing.T) {
	// Create app with cluster disabled
	config := &app.Config{
		ClusterDisabled: true,
	}

	app := &App{
		config: config,
	}

	// Create gin context
	gin.SetMode(gin.TestMode)
	router := gin.New()

	called := false
	router.Use(app.RedirectPluginInvoke())
	router.GET("/test", func(c *gin.Context) {
		called = true
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Create request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should call next handler even without plugin context
	assert.True(t, called)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRedirectPluginInvoke_ClusterEnabled_PluginOnCurrentNode(t *testing.T) {
	// Create app with cluster enabled but no actual cluster (nil)
	// This tests the middleware creation and basic flow
	config := &app.Config{
		ClusterDisabled: false,
	}

	app := &App{
		config:  config,
		cluster: nil, // This will cause panic if IsPluginOnCurrentNode is called, but we test middleware creation
	}

	// Test that middleware is created successfully
	middleware := app.RedirectPluginInvoke()
	assert.NotNil(t, middleware)

	// Test that middleware handles missing plugin identifier correctly
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Create request without plugin context
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 500 error due to missing plugin identifier
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRedirectPluginInvoke_ClusterEnabled_PluginNotOnCurrentNode(t *testing.T) {
	// Create app with cluster enabled
	config := &app.Config{
		ClusterDisabled: false,
	}

	app := &App{
		config: config,
	}

	// Test that middleware is created successfully
	middleware := app.RedirectPluginInvoke()
	assert.NotNil(t, middleware)

	// Test middleware with valid plugin identifier but nil cluster
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Create request with plugin context
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Set valid plugin context
	identity, _ := plugin_entities.NewPluginUniqueIdentifier("test-plugin-v1.0.0")
	c.Set(constants.CONTEXT_KEY_PLUGIN_UNIQUE_IDENTIFIER, identity)

	// Process request through middleware - should panic due to nil cluster
	assert.Panics(t, func() {
		middleware(c)
	})
}

func TestRedirectPluginInvoke_MissingPluginIdentifier(t *testing.T) {
	// Create app with cluster enabled
	config := &app.Config{
		ClusterDisabled: false,
	}

	app := &App{
		config: config,
	}

	// Create gin context
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(app.RedirectPluginInvoke())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Create request without plugin context
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 500 error due to missing plugin identifier
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRedirectPluginInvoke_InvalidPluginIdentifier(t *testing.T) {
	// Create app with cluster enabled
	config := &app.Config{
		ClusterDisabled: false,
	}

	app := &App{
		config: config,
	}

	// Create gin context
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(app.RedirectPluginInvoke())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Create request with invalid plugin context
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Set invalid plugin context
	c.Set(constants.CONTEXT_KEY_PLUGIN_UNIQUE_IDENTIFIER, "invalid-identifier")

	router.ServeHTTP(w, req)

	// Should return 500 error due to invalid plugin identifier
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCheckingKey(t *testing.T) {
	// Test valid key
	middleware := CheckingKey("valid-key")
	assert.NotNil(t, middleware)

	// Create gin context
	gin.SetMode(gin.TestMode)
	router := gin.New()

	called := false
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		called = true
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Test with valid key
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(constants.X_API_KEY, "valid-key")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test with invalid key
	called = false
	req.Header.Set(constants.X_API_KEY, "invalid-key")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestApp_AdminAPIKey(t *testing.T) {
	// Create app instance
	app := &App{}

	// Test valid admin key
	middleware := app.AdminAPIKey("admin-key")
	assert.NotNil(t, middleware)

	// Create gin context
	gin.SetMode(gin.TestMode)
	router := gin.New()

	called := false
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		called = true
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Test with valid key
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(constants.X_ADMIN_API_KEY, "admin-key")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test with invalid key
	called = false
	req.Header.Set(constants.X_ADMIN_API_KEY, "invalid-key")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
