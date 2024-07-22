package service

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/tool_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

func InvokeTool(r *plugin_entities.InvokePluginRequest[requests.RequestInvokeTool], ctx *gin.Context) {
	// create session
	session := createSession(r)
	defer session.Close()

	baseSSEService(r, func() (*stream.StreamResponse[tool_entities.ToolResponseChunk], error) {
		return plugin_daemon.InvokeTool(session, &r.Data)
	}, ctx)
}

func ValidateToolCredentials(r *plugin_entities.InvokePluginRequest[requests.RequestValidateToolCredentials], ctx *gin.Context) {
	// create session
	session := createSession(r)
	defer session.Close()

	baseSSEService(r, func() (*stream.StreamResponse[tool_entities.ValidateCredentialsResult], error) {
		return plugin_daemon.ValidateToolCredentials(session, &r.Data)
	}, ctx)
}
