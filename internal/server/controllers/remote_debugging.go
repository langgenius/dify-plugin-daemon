package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/service"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/requests"
)

func GetRemoteDebuggingKey(c *gin.Context) {
	BindRequest[requests.RequestGetRemoteDebuggingKey](
		c, func(request requests.RequestGetRemoteDebuggingKey) {
			c.JSON(200, service.GetRemoteDebuggingKey(request.TenantID))
		},
	)
}