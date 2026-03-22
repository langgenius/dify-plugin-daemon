package service

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/backwards_invocation/transaction"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

func HandleServerlessPluginTransaction(handler *transaction.ServerlessTransactionHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		// get session id from the context
		sessionId := c.Request.Header.Get("Dify-Plugin-Session-ID")
		if sessionId == "" {
			log.WarnContext(
				c.Request.Context(),
				"missing Dify-Plugin-Session-ID header",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
			)
		}

		handler.Handle(c, sessionId)
	}
}
