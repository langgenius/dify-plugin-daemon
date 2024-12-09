package service

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/agent_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

func InvokeAgent(
	r *plugin_entities.InvokePluginRequest[requests.RequestInvokeAgent],
	ctx *gin.Context,
	max_timeout_seconds int,
) {
	// create session
	session, err := createSession(
		r,
		access_types.PLUGIN_ACCESS_TYPE_AGENT,
		access_types.PLUGIN_ACCESS_ACTION_INVOKE_AGENT,
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
		func() (*stream.Stream[agent_entities.AgentResponseChunk], error) {
			return plugin_daemon.InvokeAgent(session, &r.Data)
		},
		ctx,
		max_timeout_seconds,
	)
}
