package service

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

func createSession[T any](
	r *plugin_entities.InvokePluginRequest[T],
	access_type access_types.PluginAccessType,
	access_action access_types.PluginAccessAction,
	cluster_id string,
) (*session_manager.Session, error) {
	runtime := plugin_manager.GetGlobalPluginManager().Get(r.PluginUniqueIdentifier)

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

func InvokeLLM(
	r *plugin_entities.InvokePluginRequest[requests.RequestInvokeLLM],
	ctx *gin.Context,
	max_timeout_seconds int,
) {
	// create session
	session, err := createSession(
		r,
		access_types.PLUGIN_ACCESS_TYPE_MODEL,
		access_types.PLUGIN_ACCESS_ACTION_INVOKE_LLM,
		ctx.GetString("cluster_id"),
	)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer session.Close()

	baseSSEService(
		func() (*stream.StreamResponse[model_entities.LLMResultChunk], error) {
			return plugin_daemon.InvokeLLM(session, &r.Data)
		},
		ctx,
		max_timeout_seconds,
	)
}

func InvokeTextEmbedding(
	r *plugin_entities.InvokePluginRequest[requests.RequestInvokeTextEmbedding],
	ctx *gin.Context,
	max_timeout_seconds int,
) {
	// create session
	session, err := createSession(
		r,
		access_types.PLUGIN_ACCESS_TYPE_MODEL,
		access_types.PLUGIN_ACCESS_ACTION_INVOKE_TEXT_EMBEDDING,
		ctx.GetString("cluster_id"))
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer session.Close()

	baseSSEService(
		func() (*stream.StreamResponse[model_entities.TextEmbeddingResult], error) {
			return plugin_daemon.InvokeTextEmbedding(session, &r.Data)
		},
		ctx,
		max_timeout_seconds,
	)
}

func InvokeRerank(
	r *plugin_entities.InvokePluginRequest[requests.RequestInvokeRerank],
	ctx *gin.Context,
	max_timeout_seconds int,
) {
	// create session
	session, err := createSession(
		r,
		access_types.PLUGIN_ACCESS_TYPE_MODEL,
		access_types.PLUGIN_ACCESS_ACTION_INVOKE_RERANK,
		ctx.GetString("cluster_id"),
	)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer session.Close()

	baseSSEService(
		func() (*stream.StreamResponse[model_entities.RerankResult], error) {
			return plugin_daemon.InvokeRerank(session, &r.Data)
		},
		ctx,
		max_timeout_seconds,
	)
}

func InvokeTTS(
	r *plugin_entities.InvokePluginRequest[requests.RequestInvokeTTS],
	ctx *gin.Context,
	max_timeout_seconds int,
) {
	// create session
	session, err := createSession(
		r,
		access_types.PLUGIN_ACCESS_TYPE_MODEL,
		access_types.PLUGIN_ACCESS_ACTION_INVOKE_TTS,
		ctx.GetString("cluster_id"),
	)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer session.Close()

	baseSSEService(
		func() (*stream.StreamResponse[model_entities.TTSResult], error) {
			return plugin_daemon.InvokeTTS(session, &r.Data)
		},
		ctx,
		max_timeout_seconds,
	)
}

func InvokeSpeech2Text(
	r *plugin_entities.InvokePluginRequest[requests.RequestInvokeSpeech2Text],
	ctx *gin.Context,
	max_timeout_seconds int,
) {
	// create session
	session, err := createSession(
		r,
		access_types.PLUGIN_ACCESS_TYPE_MODEL,
		access_types.PLUGIN_ACCESS_ACTION_INVOKE_SPEECH2TEXT,
		ctx.GetString("cluster_id"),
	)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer session.Close()

	baseSSEService(
		func() (*stream.StreamResponse[model_entities.Speech2TextResult], error) {
			return plugin_daemon.InvokeSpeech2Text(session, &r.Data)
		},
		ctx,
		max_timeout_seconds,
	)
}

func InvokeModeration(
	r *plugin_entities.InvokePluginRequest[requests.RequestInvokeModeration],
	ctx *gin.Context,
	max_timeout_seconds int,
) {
	// create session
	session, err := createSession(
		r,
		access_types.PLUGIN_ACCESS_TYPE_MODEL,
		access_types.PLUGIN_ACCESS_ACTION_INVOKE_MODERATION,
		ctx.GetString("cluster_id"),
	)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer session.Close()

	baseSSEService(
		func() (*stream.StreamResponse[model_entities.ModerationResult], error) {
			return plugin_daemon.InvokeModeration(session, &r.Data)
		},
		ctx,
		max_timeout_seconds,
	)
}

func ValidateProviderCredentials(
	r *plugin_entities.InvokePluginRequest[requests.RequestValidateProviderCredentials],
	ctx *gin.Context,
	max_timeout_seconds int,
) {
	// create session
	session, err := createSession(
		r,
		access_types.PLUGIN_ACCESS_TYPE_MODEL,
		access_types.PLUGIN_ACCESS_ACTION_VALIDATE_PROVIDER_CREDENTIALS,
		ctx.GetString("cluster_id"),
	)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer session.Close()

	baseSSEService(
		func() (*stream.StreamResponse[model_entities.ValidateCredentialsResult], error) {
			return plugin_daemon.ValidateProviderCredentials(session, &r.Data)
		},
		ctx,
		max_timeout_seconds,
	)
}

func ValidateModelCredentials(
	r *plugin_entities.InvokePluginRequest[requests.RequestValidateModelCredentials],
	ctx *gin.Context,
	max_timeout_seconds int,
) {
	// create session
	session, err := createSession(
		r,
		access_types.PLUGIN_ACCESS_TYPE_MODEL,
		access_types.PLUGIN_ACCESS_ACTION_VALIDATE_MODEL_CREDENTIALS,
		ctx.GetString("cluster_id"),
	)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer session.Close()

	baseSSEService(
		func() (*stream.StreamResponse[model_entities.ValidateCredentialsResult], error) {
			return plugin_daemon.ValidateModelCredentials(session, &r.Data)
		},
		ctx,
		max_timeout_seconds,
	)
}
