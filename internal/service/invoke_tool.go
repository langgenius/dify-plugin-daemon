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

func InvokeLLM(r *plugin_entities.InvokePluginRequest[requests.RequestInvokeLLM], ctx *gin.Context) {
	// create session
	session := session_manager.NewSession(r.TenantId, r.UserId, parser.MarshalPluginIdentity(r.PluginName, r.PluginVersion))
	defer session.Close()

	baseSSEService(r, func() (*stream.StreamResponse[model_entities.LLMResultChunk], error) {
		return plugin_daemon.InvokeLLM(session, &r.Data)
	}, ctx)
}

func InvokeTextEmbedding(r *plugin_entities.InvokePluginRequest[requests.RequestInvokeTextEmbedding], ctx *gin.Context) {
	// create session
	session := session_manager.NewSession(r.TenantId, r.UserId, parser.MarshalPluginIdentity(r.PluginName, r.PluginVersion))
	defer session.Close()

	baseSSEService(r, func() (*stream.StreamResponse[model_entities.TextEmbeddingResult], error) {
		return plugin_daemon.InvokeTextEmbedding(session, &r.Data)
	}, ctx)
}

func InvokeRerank(r *plugin_entities.InvokePluginRequest[requests.RequestInvokeRerank], ctx *gin.Context) {
	// create session
	session := session_manager.NewSession(r.TenantId, r.UserId, parser.MarshalPluginIdentity(r.PluginName, r.PluginVersion))
	defer session.Close()

	baseSSEService(r, func() (*stream.StreamResponse[model_entities.RerankResult], error) {
		return plugin_daemon.InvokeRerank(session, &r.Data)
	}, ctx)
}

func InvokeTTS(r *plugin_entities.InvokePluginRequest[requests.RequestInvokeTTS], ctx *gin.Context) {
	// create session
	session := session_manager.NewSession(r.TenantId, r.UserId, parser.MarshalPluginIdentity(r.PluginName, r.PluginVersion))
	defer session.Close()

	baseSSEService(r, func() (*stream.StreamResponse[model_entities.TTSResult], error) {
		return plugin_daemon.InvokeTTS(session, &r.Data)
	}, ctx)
}

func InvokeSpeech2Text(r *plugin_entities.InvokePluginRequest[requests.RequestInvokeSpeech2Text], ctx *gin.Context) {
	// create session
	session := session_manager.NewSession(r.TenantId, r.UserId, parser.MarshalPluginIdentity(r.PluginName, r.PluginVersion))
	defer session.Close()

	baseSSEService(r, func() (*stream.StreamResponse[model_entities.Speech2TextResult], error) {
		return plugin_daemon.InvokeSpeech2Text(session, &r.Data)
	}, ctx)
}

func InvokeModeration(r *plugin_entities.InvokePluginRequest[requests.RequestInvokeModeration], ctx *gin.Context) {
	// create session
	session := session_manager.NewSession(r.TenantId, r.UserId, parser.MarshalPluginIdentity(r.PluginName, r.PluginVersion))
	defer session.Close()

	baseSSEService(r, func() (*stream.StreamResponse[model_entities.ModerationResult], error) {
		return plugin_daemon.InvokeModeration(session, &r.Data)
	}, ctx)
}

func ValidateProviderCredentials(r *plugin_entities.InvokePluginRequest[requests.RequestValidateProviderCredentials], ctx *gin.Context) {
	// create session
	session := session_manager.NewSession(r.TenantId, r.UserId, parser.MarshalPluginIdentity(r.PluginName, r.PluginVersion))
	defer session.Close()

	baseSSEService(r, func() (*stream.StreamResponse[model_entities.ValidateCredentialsResult], error) {
		return plugin_daemon.ValidateProviderCredentials(session, &r.Data)
	}, ctx)
}

func ValidateModelCredentials(r *plugin_entities.InvokePluginRequest[requests.RequestValidateModelCredentials], ctx *gin.Context) {
	// create session
	session := session_manager.NewSession(r.TenantId, r.UserId, parser.MarshalPluginIdentity(r.PluginName, r.PluginVersion))
	defer session.Close()

	baseSSEService(r, func() (*stream.StreamResponse[model_entities.ValidateCredentialsResult], error) {
		return plugin_daemon.ValidateModelCredentials(session, &r.Data)
	}, ctx)
}
