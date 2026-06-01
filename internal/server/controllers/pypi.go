package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/pip"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
)

// ListPypiMirrors returns the candidate PyPI mirrors together with the latest
// probe latency/reachability and which one is currently selected.
func ListPypiMirrors(c *gin.Context) {
	provider := pip.Provider()
	if provider == nil {
		c.JSON(http.StatusOK, exception.InternalServerError(
			errors.New("pypi provider is not initialized"),
		).ToResponse())
		return
	}

	c.JSON(http.StatusOK, entities.NewSuccessResponse(pip.BuildMirrorListing(provider)))
}

// SelectPypiMirror selects (and persists) the effective PyPI mirror. An empty
// or unknown URL is materialized as a custom mirror in the database.
func SelectPypiMirror(c *gin.Context) {
	BindRequest(c, func(request struct {
		URL  string `json:"url" validate:"required"`
		Name string `json:"name"`
	}) {
		provider := pip.Provider()
		mutable, ok := provider.(pip.MutableProvider)
		if !ok {
			c.JSON(http.StatusOK, exception.BadRequestError(
				errors.New("mirror selection is not supported (no database-backed provider)"),
			).ToResponse())
			return
		}

		if err := mutable.Select(pip.Mirror{Name: request.Name, URL: request.URL}); err != nil {
			c.JSON(http.StatusOK, exception.InternalServerError(err).ToResponse())
			return
		}

		c.JSON(http.StatusOK, entities.NewSuccessResponse(pip.BuildMirrorListing(provider)))
	})
}
