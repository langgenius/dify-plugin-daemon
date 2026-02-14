package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

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
