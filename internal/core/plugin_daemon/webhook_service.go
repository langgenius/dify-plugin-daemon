package plugin_daemon

import (
	"encoding/hex"
	"net/http"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/webhook_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

func InvokeWebhook(
	session *session_manager.Session,
	request *requests.RequestInvokeWebhook,
) (
	int, *http.Header, *stream.StreamResponse[[]byte], error,
) {
	resp, err := genericInvokePlugin[requests.RequestInvokeWebhook, webhook_entities.WebhookResponseChunk](
		session,
		request,
		128,
		access_types.PLUGIN_ACCESS_TYPE_WEBHOOK,
		access_types.PLUGIN_ACCESS_ACTION_WEBHOOK,
	)

	if err != nil {
		return http.StatusInternalServerError, nil, nil, err
	}

	status_code := http.StatusContinue
	headers := &http.Header{}
	response := stream.NewStreamResponse[[]byte](128)
	response.OnClose(func() {
		// add close callback, ensure resources are released
		resp.Close()
	})

	for resp.Next() {
		result, err := resp.Read()
		if err != nil {
			resp.Close()
			return http.StatusInternalServerError, nil, nil, err
		}

		if result.Status != nil {
			status_code = int(*result.Status)
		}

		if result.Headers != nil {
			for k, v := range result.Headers {
				headers.Add(k, v)
			}
		}

		if result.Result != nil {
			dehexed, err := hex.DecodeString(*result.Result)
			if err != nil {
				resp.Close()
				return http.StatusInternalServerError, nil, nil, err
			}

			response.Write(dehexed)
			routine.Submit(func() {
				defer response.Close()
				for resp.Next() {
					chunk, err := resp.Read()
					if err != nil {
						return
					}

					dehexed, err := hex.DecodeString(*chunk.Result)
					if err != nil {
						return
					}
					response.Write(dehexed)
				}
			})
			break
		}
	}

	return status_code, headers, response, nil
}
