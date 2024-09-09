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
	var err error

	context_type := r.GetHeader("Content-Type")
	if context_type == "application/json" {
		err = r.ShouldBindJSON(&request)
	} else {
		err = r.ShouldBind(&request)
	}

	if err != nil {
		resp := entities.NewErrorResponse(-400, err.Error())
		r.JSON(400, resp)
		return
	}

	// validate, we have customized some validators which are not supported by gin binding
	if err := validators.GlobalEntitiesValidator.Struct(request); err != nil {
		resp := entities.NewErrorResponse(-400, err.Error())
		r.JSON(400, resp)
		return
	}

	success(request)
}

func BindRequestWithPluginUniqueIdentifier[T any](r *gin.Context, success func(T, plugin_entities.PluginUniqueIdentifier)) {
	BindRequest(r, func(req T) {
		plugin_unique_identifier := r.GetHeader(constants.X_PLUGIN_IDENTIFIER)
		if plugin_unique_identifier == "" {
			resp := entities.NewErrorResponse(-400, "Plugin unique identifier is required")
			r.JSON(400, resp)
			return
		}

		success(req, plugin_entities.PluginUniqueIdentifier(plugin_unique_identifier))
	})
}
