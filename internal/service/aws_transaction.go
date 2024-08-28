package service

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/backwards_invocation/transaction"
)

func HandleAWSPluginTransaction(handler *transaction.AWSTransactionHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		// get session id from the context
		session_id := c.Request.Header.Get("Dify-Plugin-Session-ID")

		handler.Handle(c, session_id)
	}
}
