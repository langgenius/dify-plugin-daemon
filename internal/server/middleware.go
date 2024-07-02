package server

import "github.com/gin-gonic/gin"

func CheckingKey(key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// get header X-Api-Key
		if c.GetHeader("X-Api-Key") != key {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		c.Next()
	}
}
