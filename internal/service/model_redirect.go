package service

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel"
	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/stream"
)

func StreamModelRedirect[R any](
	request *plugin_entities.InvokePluginRequest[io_tunnel.ModelDispatchData],
	ctx *gin.Context,
	action access_types.PluginAccessAction,
	bufferSize int,
	timeout int,
) {
	baseSSEWithSession(
		func(session *session_manager.Session) (*stream.Stream[R], error) {
			return io_tunnel.GenericInvokePlugin[io_tunnel.ModelDispatchData, R](session, &request.Data, bufferSize)
		},
		access_types.PLUGIN_ACCESS_TYPE_MODEL, action, request, ctx, timeout,
	)
}

func PollModelRedirect(
	request *plugin_entities.InvokePluginRequest[io_tunnel.ModelDispatchData],
	ctx *gin.Context,
	action access_types.PluginAccessAction,
) {
	session, err := createSession(request, access_types.PLUGIN_ACCESS_TYPE_MODEL, action, ctx.GetString("cluster_id"), ctx.Request.Context())
	if err != nil {
		ctx.JSON(500, exception.InternalServerError(err).ToResponse())
		return
	}
	defer session.Close(session_manager.CloseSessionPayload{IgnoreCache: false})

	response, err := io_tunnel.PollModelRedirect(session, &request.Data)
	if err != nil {
		ctx.JSON(500, exception.InvokePluginError(err).ToResponse())
		return
	}
	ctx.JSON(200, entities.NewSuccessResponse(response))
}
