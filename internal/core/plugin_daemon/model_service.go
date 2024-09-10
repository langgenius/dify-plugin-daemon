package plugin_daemon

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

func InvokeLLM(
	session *session_manager.Session,
	request *requests.RequestInvokeLLM,
) (
	*stream.Stream[model_entities.LLMResultChunk], error,
) {
	return genericInvokePlugin[requests.RequestInvokeLLM, model_entities.LLMResultChunk](
		session,
		request,
		512,
	)
}

func InvokeTextEmbedding(
	session *session_manager.Session,
	request *requests.RequestInvokeTextEmbedding,
) (
	*stream.Stream[model_entities.TextEmbeddingResult], error,
) {
	return genericInvokePlugin[requests.RequestInvokeTextEmbedding, model_entities.TextEmbeddingResult](
		session,
		request,
		1,
	)
}

func InvokeRerank(
	session *session_manager.Session,
	request *requests.RequestInvokeRerank,
) (
	*stream.Stream[model_entities.RerankResult], error,
) {
	return genericInvokePlugin[requests.RequestInvokeRerank, model_entities.RerankResult](
		session,
		request,
		1,
	)
}

func InvokeTTS(
	session *session_manager.Session,
	request *requests.RequestInvokeTTS,
) (
	*stream.Stream[model_entities.TTSResult], error,
) {
	return genericInvokePlugin[requests.RequestInvokeTTS, model_entities.TTSResult](
		session,
		request,
		1,
	)
}

func InvokeSpeech2Text(
	session *session_manager.Session,
	request *requests.RequestInvokeSpeech2Text,
) (
	*stream.Stream[model_entities.Speech2TextResult], error,
) {
	return genericInvokePlugin[requests.RequestInvokeSpeech2Text, model_entities.Speech2TextResult](
		session,
		request,
		1,
	)
}

func InvokeModeration(
	session *session_manager.Session,
	request *requests.RequestInvokeModeration,
) (
	*stream.Stream[model_entities.ModerationResult], error,
) {
	return genericInvokePlugin[requests.RequestInvokeModeration, model_entities.ModerationResult](
		session,
		request,
		1,
	)
}

func ValidateProviderCredentials(
	session *session_manager.Session,
	request *requests.RequestValidateProviderCredentials,
) (
	*stream.Stream[model_entities.ValidateCredentialsResult], error,
) {
	return genericInvokePlugin[requests.RequestValidateProviderCredentials, model_entities.ValidateCredentialsResult](
		session,
		request,
		1,
	)
}

func ValidateModelCredentials(
	session *session_manager.Session,
	request *requests.RequestValidateModelCredentials,
) (
	*stream.Stream[model_entities.ValidateCredentialsResult], error,
) {
	return genericInvokePlugin[requests.RequestValidateModelCredentials, model_entities.ValidateCredentialsResult](
		session,
		request,
		1,
	)
}
