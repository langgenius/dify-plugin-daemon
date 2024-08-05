package service

import (
	"bytes"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
)

func Webhook(ctx *gin.Context, webhook *models.Webhook, path string) {
	req := ctx.Request

	var buffer bytes.Buffer
	err := req.Write(&buffer)

	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
	}

	// fetch plugin

}
