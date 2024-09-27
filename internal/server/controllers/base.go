package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/server/constants"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/validators"
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
		resp := entities.NewErrorResponse(-400, err.Error())
		r.JSON(400, resp)
		return
	}

	success(request)
}

func BindPluginDispatchRequest[T any](r *gin.Context, success func(
	plugin_entities.InvokePluginRequest[T],
)) {
	BindRequest(r, func(req plugin_entities.InvokePluginRequest[T]) {
		plugin_unique_identifier_any, exists := r.Get(constants.CONTEXT_KEY_PLUGIN_UNIQUE_IDENTIFIER)
		if !exists {
			resp := entities.NewErrorResponse(-400, "Plugin unique identifier is required")
			r.JSON(400, resp)
			return
		}

		plugin_unique_identifier, ok := plugin_unique_identifier_any.(plugin_entities.PluginUniqueIdentifier)
		if !ok {
			resp := entities.NewErrorResponse(-400, "Plugin unique identifier is required")
			r.JSON(400, resp)
			return
		}

		// set plugin unique identifier
		req.UniqueIdentifier = plugin_unique_identifier

		success(req)
	})
}
