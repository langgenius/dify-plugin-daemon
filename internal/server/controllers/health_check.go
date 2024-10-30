package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

func HealthCheck(c *gin.Context) {
	routine.InitPool(10)
	c.JSON(200, gin.H{"status": "ok", "pool_status": routine.FetchRoutineStatus()})
}
