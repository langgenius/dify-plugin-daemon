package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
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

	// validate
	if err := validators.GlobalEntitiesValidator.Struct(request); err != nil {
		resp := entities.NewErrorResponse(-400, "Invalid request")
		r.JSON(400, resp)
		return
	}

	if err != nil {
		resp := entities.NewErrorResponse(-400, "Invalid request")
		r.JSON(400, resp)
		return
	}
	success(request)
}
