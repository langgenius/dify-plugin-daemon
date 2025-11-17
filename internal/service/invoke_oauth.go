package service

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel"
	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/oauth_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/stream"
)

func OAuthGetAuthorizationURL(
	r *plugin_entities.InvokePluginRequest[requests.RequestOAuthGetAuthorizationURL],
	ctx *gin.Context,
	maxExecutionTimeout time.Duration,
) {
	baseSSEWithSession(
		func(session *session_manager.Session) (*stream.Stream[oauth_entities.OAuthGetAuthorizationURLResult], error) {
			return io_tunnel.OAuthGetAuthorizationURL(session, &r.Data)
		},
		access_types.PLUGIN_ACCESS_TYPE_OAUTH,
		access_types.PLUGIN_ACCESS_ACTION_GET_AUTHORIZATION_URL,
		r,
		ctx,
		int(maxExecutionTimeout.Seconds()),
	)
}

func OAuthGetCredentials(
	r *plugin_entities.InvokePluginRequest[requests.RequestOAuthGetCredentials],
	ctx *gin.Context,
	maxExecutionTimeout time.Duration,
) {
	baseSSEWithSession(
		func(session *session_manager.Session) (*stream.Stream[oauth_entities.OAuthGetCredentialsResult], error) {
			return io_tunnel.OAuthGetCredentials(session, &r.Data)
		},
		access_types.PLUGIN_ACCESS_TYPE_OAUTH,
		access_types.PLUGIN_ACCESS_ACTION_GET_CREDENTIALS,
		r,
		ctx,
		int(maxExecutionTimeout.Seconds()),
	)
}
