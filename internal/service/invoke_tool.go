package service

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/tool_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

func createSession[T any](
	r *plugin_entities.InvokePluginRequest[T],
	access_type access_types.PluginAccessType,
	access_action access_types.PluginAccessAction,
	cluster_id string,
) (*session_manager.Session, error) {
	runtime := plugin_manager.Manager().Get(r.PluginUniqueIdentifier)

	session := session_manager.NewSession(
		r.TenantId,
		r.UserId,
		r.PluginUniqueIdentifier,
		cluster_id,
		access_type,
		access_action,
		runtime.Configuration(),
	)

	session.BindRuntime(runtime)
	return session, nil
}

func InvokeTool(
	r *plugin_entities.InvokePluginRequest[requests.RequestInvokeTool],
	ctx *gin.Context,
	max_timeout_seconds int,
) {
	// create session
	session, err := createSession(
		r,
		access_types.PLUGIN_ACCESS_TYPE_TOOL,
		access_types.PLUGIN_ACCESS_ACTION_INVOKE_TOOL,
		ctx.GetString("cluster_id"),
	)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer session.Close()

	baseSSEService(
		func() (*stream.Stream[tool_entities.ToolResponseChunk], error) {
			return plugin_daemon.InvokeTool(session, &r.Data)
		},
		ctx,
		max_timeout_seconds,
	)
}

func ValidateToolCredentials(
	r *plugin_entities.InvokePluginRequest[requests.RequestValidateToolCredentials],
	ctx *gin.Context,
	max_timeout_seconds int,
) {
	// create session
	session, err := createSession(
		r,
		access_types.PLUGIN_ACCESS_TYPE_TOOL,
		access_types.PLUGIN_ACCESS_ACTION_VALIDATE_TOOL_CREDENTIALS,
		ctx.GetString("cluster_id"),
	)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer session.Close()

	baseSSEService(
		func() (*stream.Stream[tool_entities.ValidateCredentialsResult], error) {
			return plugin_daemon.ValidateToolCredentials(session, &r.Data)
		},
		ctx,
		max_timeout_seconds,
	)
}
