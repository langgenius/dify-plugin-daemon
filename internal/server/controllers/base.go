package controllers

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/server/constants"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/validators"
)

func BindRequest[T any](r *gin.Context, success func(T)) {
	var request T

	if r.Request.Header.Get("Content-Type") == "application/json" {
		r.ShouldBindJSON(&request)
	} else {
		r.ShouldBind(&request)
	}

	// bind uri
	r.ShouldBindUri(&request)

	// validate, we have customized some validators which are not supported by gin binding
	if err := validators.GlobalEntitiesValidator.Struct(request); err != nil {
		r.JSON(400, exception.BadRequestError(err).ToResponse())
		return
	}

	success(request)
}

func BindPluginDispatchRequest[T any](r *gin.Context, success func(
	plugin_entities.InvokePluginRequest[T],
)) {
	BindRequest(r, func(req plugin_entities.InvokePluginRequest[T]) {
		pluginUniqueIdentifierAny, exists := r.Get(constants.CONTEXT_KEY_PLUGIN_UNIQUE_IDENTIFIER)
		if !exists {
			r.JSON(400, exception.UniqueIdentifierError(errors.New("Plugin unique identifier is required")).ToResponse())
			return
		}

		pluginUniqueIdentifier, ok := pluginUniqueIdentifierAny.(plugin_entities.PluginUniqueIdentifier)
		if !ok {
			r.JSON(400, exception.UniqueIdentifierError(errors.New("Plugin unique identifier is not valid")).ToResponse())
			return
		}

		// set plugin unique identifier
		req.UniqueIdentifier = pluginUniqueIdentifier

		success(req)
	})
}
