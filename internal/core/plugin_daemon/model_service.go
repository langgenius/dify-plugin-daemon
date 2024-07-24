package plugin_daemon

import (
	"errors"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

func genericInvokePlugin[Req any, Rsp any](
	session *session_manager.Session,
	request *Req,
	response_buffer_size int,
	typ PluginAccessType,
	action PluginAccessAction,
) (
	*stream.StreamResponse[Rsp], error,
) {
	runtime := plugin_manager.Get(session.PluginIdentity())
	if runtime == nil {
		return nil, errors.New("plugin not found")
	}

	response := stream.NewStreamResponse[Rsp](response_buffer_size)

	listener := runtime.Listen(session.ID())
	listener.AddListener(func(message []byte) {
		chunk, err := parser.UnmarshalJsonBytes[plugin_entities.SessionMessage](message)
		if err != nil {
			log.Error("unmarshal json failed: %s", err.Error())
			return
		}

		switch chunk.Type {
		case plugin_entities.SESSION_MESSAGE_TYPE_STREAM:
			chunk, err := parser.UnmarshalJsonBytes[Rsp](chunk.Data)
			if err != nil {
				log.Error("unmarshal json failed: %s", err.Error())
				return
			}
			response.Write(chunk)
		case plugin_entities.SESSION_MESSAGE_TYPE_INVOKE:
			invokeDify(runtime, typ, session, chunk.Data)
		case plugin_entities.SESSION_MESSAGE_TYPE_END:
			response.Close()
		case plugin_entities.SESSION_MESSAGE_TYPE_ERROR:
			e, err := parser.UnmarshalJsonBytes[plugin_entities.ErrorResponse](chunk.Data)
			if err != nil {
				break
			}
			response.WriteError(errors.New(e.Error))
			response.Close()
		default:
			response.WriteError(errors.New("unknown stream message type: " + string(chunk.Type)))
			response.Close()
		}
	})

	response.OnClose(func() {
		listener.Close()
	})

	session.Write(
		session_manager.PLUGIN_IN_STREAM_EVENT_REQUEST,
		getInvokeModelMap(
			session,
			typ,
			action,
			request,
		),
	)

	return response, nil
}

func getInvokeModelMap(
	session *session_manager.Session,
	typ PluginAccessType,
	action PluginAccessAction,
	request any,
) map[string]any {
	req := getBasicPluginAccessMap(session.UserID(), typ, action)
	for k, v := range parser.StructToMap(request) {
		req[k] = v
	}
	return req
}

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
		PLUGIN_ACCESS_TYPE_MODEL,
		PLUGIN_ACCESS_ACTION_INVOKE_LLM,
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
		PLUGIN_ACCESS_TYPE_MODEL,
		PLUGIN_ACCESS_ACTION_INVOKE_TEXT_EMBEDDING,
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
		PLUGIN_ACCESS_TYPE_MODEL,
		PLUGIN_ACCESS_ACTION_INVOKE_RERANK,
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
		PLUGIN_ACCESS_TYPE_MODEL,
		PLUGIN_ACCESS_ACTION_INVOKE_TTS,
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
		PLUGIN_ACCESS_TYPE_MODEL,
		PLUGIN_ACCESS_ACTION_INVOKE_SPEECH2TEXT,
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
		PLUGIN_ACCESS_TYPE_MODEL,
		PLUGIN_ACCESS_ACTION_INVOKE_MODERATION,
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
		PLUGIN_ACCESS_TYPE_MODEL,
		PLUGIN_ACCESS_ACTION_VALIDATE_PROVIDER_CREDENTIALS,
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
		PLUGIN_ACCESS_TYPE_MODEL,
		PLUGIN_ACCESS_ACTION_VALIDATE_MODEL_CREDENTIALS,
	)
}
