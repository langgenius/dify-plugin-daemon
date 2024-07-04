package server

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
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
		resp := entities.NewErrorResponse(-400, "Invalid request")
		r.JSON(400, resp)
		return
	}
	success(request)
}
