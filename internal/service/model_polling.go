package service

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel"
	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
)

func StartPolling(
	r *plugin_entities.InvokePluginRequest[requests.RequestStartPolling],
	ctx *gin.Context,
) {
	session, err := createSession(
		r,
		access_types.PLUGIN_ACCESS_TYPE_MODEL,
		access_types.PLUGIN_ACCESS_ACTION_START_POLLING,
		ctx.GetString("cluster_id"),
		ctx.Request.Context(),
	)
	if err != nil {
		ctx.JSON(500, exception.InternalServerError(err).ToResponse())
		return
	}
	defer session.Close(session_manager.CloseSessionPayload{IgnoreCache: false})

	resp, err := io_tunnel.StartPolling(session, &r.Data)
	if err != nil {
		ctx.JSON(500, exception.InvokePluginError(err).ToResponse())
		return
	}
	ctx.JSON(200, entities.NewSuccessResponse(resp))
}

func CheckPolling(
	r *plugin_entities.InvokePluginRequest[requests.RequestCheckPolling],
	ctx *gin.Context,
) {
	session, err := createSession(
		r,
		access_types.PLUGIN_ACCESS_TYPE_MODEL,
		access_types.PLUGIN_ACCESS_ACTION_CHECK_POLLING,
		ctx.GetString("cluster_id"),
		ctx.Request.Context(),
	)
	if err != nil {
		ctx.JSON(500, exception.InternalServerError(err).ToResponse())
		return
	}
	defer session.Close(session_manager.CloseSessionPayload{IgnoreCache: false})

	resp, err := io_tunnel.CheckPolling(session, &r.Data)
	if err != nil {
		ctx.JSON(500, exception.InvokePluginError(err).ToResponse())
		return
	}

	ctx.JSON(200, entities.NewSuccessResponse(resp))
}
