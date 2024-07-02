package server

import "github.com/gin-gonic/gin"

func InvokePlugin(c *gin.Context) {

}

func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}
