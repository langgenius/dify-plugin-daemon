package plugin_daemon

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
)

func InvokeLLM(
	session *session_manager.Session,
	request *requests.RequestInvokeLLM,
) (
	*stream.Stream[model_entities.LLMResultChunk], error,
) {
	return GenericInvokePlugin[requests.RequestInvokeLLM, model_entities.LLMResultChunk](
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
	return GenericInvokePlugin[requests.RequestInvokeTextEmbedding, model_entities.TextEmbeddingResult](
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
	return GenericInvokePlugin[requests.RequestInvokeRerank, model_entities.RerankResult](
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
	return GenericInvokePlugin[requests.RequestInvokeTTS, model_entities.TTSResult](
		session,
		request,
		512,
	)
}

func InvokeSpeech2Text(
	session *session_manager.Session,
	request *requests.RequestInvokeSpeech2Text,
) (
	*stream.Stream[model_entities.Speech2TextResult], error,
) {
	return GenericInvokePlugin[requests.RequestInvokeSpeech2Text, model_entities.Speech2TextResult](
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
	return GenericInvokePlugin[requests.RequestInvokeModeration, model_entities.ModerationResult](
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
	return GenericInvokePlugin[requests.RequestValidateProviderCredentials, model_entities.ValidateCredentialsResult](
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
	return GenericInvokePlugin[requests.RequestValidateModelCredentials, model_entities.ValidateCredentialsResult](
		session,
		request,
		1,
	)
}

func GetTTSModelVoices(
	session *session_manager.Session,
	request *requests.RequestGetTTSModelVoices,
) (
	*stream.Stream[model_entities.GetTTSVoicesResponse], error,
) {
	return GenericInvokePlugin[requests.RequestGetTTSModelVoices, model_entities.GetTTSVoicesResponse](
		session,
		request,
		1,
	)
}

func GetTextEmbeddingNumTokens(
	session *session_manager.Session,
	request *requests.RequestGetTextEmbeddingNumTokens,
) (
	*stream.Stream[model_entities.GetTextEmbeddingNumTokensResponse], error,
) {
	return GenericInvokePlugin[requests.RequestGetTextEmbeddingNumTokens, model_entities.GetTextEmbeddingNumTokensResponse](
		session,
		request,
		1,
	)
}

func GetLLMNumTokens(
	session *session_manager.Session,
	request *requests.RequestGetLLMNumTokens,
) (
	*stream.Stream[model_entities.LLMGetNumTokensResponse], error,
) {
	return GenericInvokePlugin[requests.RequestGetLLMNumTokens, model_entities.LLMGetNumTokensResponse](
		session,
		request,
		1,
	)
}

func GetAIModelSchema(
	session *session_manager.Session,
	request *requests.RequestGetAIModelSchema,
) (
	*stream.Stream[model_entities.GetModelSchemasResponse], error,
) {
	return GenericInvokePlugin[requests.RequestGetAIModelSchema, model_entities.GetModelSchemasResponse](
		session,
		request,
		1,
	)
}
