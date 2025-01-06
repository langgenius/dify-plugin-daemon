package service

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/agent_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
)

func InvokeAgentStrategy(
	r *plugin_entities.InvokePluginRequest[requests.RequestInvokeAgentStrategy],
	ctx *gin.Context,
	max_timeout_seconds int,
) {
	// create session
	session, err := createSession(
		r,
		access_types.PLUGIN_ACCESS_TYPE_AGENT_STRATEGY,
		access_types.PLUGIN_ACCESS_ACTION_INVOKE_AGENT_STRATEGY,
		ctx.GetString("cluster_id"),
	)
	if err != nil {
		ctx.JSON(500, exception.InternalServerError(err).ToResponse())
		return
	}
	defer session.Close(session_manager.CloseSessionPayload{
		IgnoreCache: false,
	})

	baseSSEService(
		func() (*stream.Stream[agent_entities.AgentStrategyResponseChunk], error) {
			return plugin_daemon.InvokeAgentStrategy(session, &r.Data)
		},
		ctx,
		max_timeout_seconds,
	)
}
