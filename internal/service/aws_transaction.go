package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/backwards_invocation/transaction"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
)

func HandleAWSPluginTransaction(handler *transaction.AWSTransactionHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		// get session id from the context
		session_id := c.GetString("session_id")
		session := session_manager.GetSession(session_id)
		if session == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "session not found"})
			return
		}

		handler.Handle(c, session_id)
	}
}
