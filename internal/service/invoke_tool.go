package service

import (
	"errors"

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
	manager := plugin_manager.Manager()
	if manager == nil {
		return nil, errors.New("failed to get plugin manager")
	}

	// try fetch plugin identifier from plugin id

	runtime := manager.Get(r.UniqueIdentifier)
	if runtime == nil {
		return nil, errors.New("failed to get plugin runtime")
	}

	session := session_manager.NewSession(
		session_manager.NewSessionPayload{
			TenantID:               r.TenantId,
			UserID:                 r.UserId,
			PluginUniqueIdentifier: r.UniqueIdentifier,
			ClusterID:              cluster_id,
			InvokeFrom:             access_type,
			Action:                 access_action,
			Declaration:            runtime.Configuration(),
			BackwardsInvocation:    manager.BackwardsInvocation(),
			IgnoreCache:            false,
		},
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
	defer session.Close(session_manager.CloseSessionPayload{
		IgnoreCache: false,
	})

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
	defer session.Close(session_manager.CloseSessionPayload{
		IgnoreCache: false,
	})

	baseSSEService(
		func() (*stream.Stream[tool_entities.ValidateCredentialsResult], error) {
			return plugin_daemon.ValidateToolCredentials(session, &r.Data)
		},
		ctx,
		max_timeout_seconds,
	)
}
