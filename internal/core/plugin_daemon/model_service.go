package plugin_daemon

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/backwards_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

func InvokeLLM(
	session *session_manager.Session,
	request *requests.RequestInvokeLLM,
) (
	*stream.StreamResponse[model_entities.LLMResultChunk], error,
) {
	return genericInvokePlugin[requests.RequestInvokeLLM, model_entities.LLMResultChunk](
		session,
		request,
		512,
		backwards_invocation.PLUGIN_ACCESS_TYPE_MODEL,
		backwards_invocation.PLUGIN_ACCESS_ACTION_INVOKE_LLM,
	)
}

func InvokeTextEmbedding(
	session *session_manager.Session,
	request *requests.RequestInvokeTextEmbedding,
) (
	*stream.StreamResponse[model_entities.TextEmbeddingResult], error,
) {
	return genericInvokePlugin[requests.RequestInvokeTextEmbedding, model_entities.TextEmbeddingResult](
		session,
		request,
		1,
		backwards_invocation.PLUGIN_ACCESS_TYPE_MODEL,
		backwards_invocation.PLUGIN_ACCESS_ACTION_INVOKE_TEXT_EMBEDDING,
	)
}

func InvokeRerank(
	session *session_manager.Session,
	request *requests.RequestInvokeRerank,
) (
	*stream.StreamResponse[model_entities.RerankResult], error,
) {
	return genericInvokePlugin[requests.RequestInvokeRerank, model_entities.RerankResult](
		session,
		request,
		1,
		backwards_invocation.PLUGIN_ACCESS_TYPE_MODEL,
		backwards_invocation.PLUGIN_ACCESS_ACTION_INVOKE_RERANK,
	)
}

func InvokeTTS(
	session *session_manager.Session,
	request *requests.RequestInvokeTTS,
) (
	*stream.StreamResponse[model_entities.TTSResult], error,
) {
	return genericInvokePlugin[requests.RequestInvokeTTS, model_entities.TTSResult](
		session,
		request,
		1,
		backwards_invocation.PLUGIN_ACCESS_TYPE_MODEL,
		backwards_invocation.PLUGIN_ACCESS_ACTION_INVOKE_TTS,
	)
}

func InvokeSpeech2Text(
	session *session_manager.Session,
	request *requests.RequestInvokeSpeech2Text,
) (
	*stream.StreamResponse[model_entities.Speech2TextResult], error,
) {
	return genericInvokePlugin[requests.RequestInvokeSpeech2Text, model_entities.Speech2TextResult](
		session,
		request,
		1,
		backwards_invocation.PLUGIN_ACCESS_TYPE_MODEL,
		backwards_invocation.PLUGIN_ACCESS_ACTION_INVOKE_SPEECH2TEXT,
	)
}

func InvokeModeration(
	session *session_manager.Session,
	request *requests.RequestInvokeModeration,
) (
	*stream.StreamResponse[model_entities.ModerationResult], error,
) {
	return genericInvokePlugin[requests.RequestInvokeModeration, model_entities.ModerationResult](
		session,
		request,
		1,
		backwards_invocation.PLUGIN_ACCESS_TYPE_MODEL,
		backwards_invocation.PLUGIN_ACCESS_ACTION_INVOKE_MODERATION,
	)
}

func ValidateProviderCredentials(
	session *session_manager.Session,
	request *requests.RequestValidateProviderCredentials,
) (
	*stream.StreamResponse[model_entities.ValidateCredentialsResult], error,
) {
	return genericInvokePlugin[requests.RequestValidateProviderCredentials, model_entities.ValidateCredentialsResult](
		session,
		request,
		1,
		backwards_invocation.PLUGIN_ACCESS_TYPE_MODEL,
		backwards_invocation.PLUGIN_ACCESS_ACTION_VALIDATE_PROVIDER_CREDENTIALS,
	)
}

func ValidateModelCredentials(
	session *session_manager.Session,
	request *requests.RequestValidateModelCredentials,
) (
	*stream.StreamResponse[model_entities.ValidateCredentialsResult], error,
) {
	return genericInvokePlugin[requests.RequestValidateModelCredentials, model_entities.ValidateCredentialsResult](
		session,
		request,
		1,
		backwards_invocation.PLUGIN_ACCESS_TYPE_MODEL,
		backwards_invocation.PLUGIN_ACCESS_ACTION_VALIDATE_MODEL_CREDENTIALS,
	)
}
