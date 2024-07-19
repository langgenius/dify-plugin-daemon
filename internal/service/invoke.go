package service

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

func InvokeTool(r *plugin_entities.InvokePluginRequest[requests.RequestInvokeTool], ctx *gin.Context) {
	// create session
	session := session_manager.NewSession(r.TenantId, r.UserId, parser.MarshalPluginIdentity(r.PluginName, r.PluginVersion))
	defer session.Close()

	baseSSEService(r, func() (*stream.StreamResponse[plugin_entities.ToolResponseChunk], error) {
		return plugin_daemon.InvokeTool(session, &r.Data)
	}, ctx)
}

func InvokeLLM(r *plugin_entities.InvokePluginRequest[requests.RequestInvokeLLM], ctx *gin.Context) {
	// create session
	session := session_manager.NewSession(r.TenantId, r.UserId, parser.MarshalPluginIdentity(r.PluginName, r.PluginVersion))
	defer session.Close()

	baseSSEService(r, func() (*stream.StreamResponse[model_entities.LLMResultChunk], error) {
		return plugin_daemon.InvokeLLM(session, &r.Data)
	}, ctx)
}
