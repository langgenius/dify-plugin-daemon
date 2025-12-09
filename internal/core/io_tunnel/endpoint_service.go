package io_tunnel

import (
	"encoding/hex"
	"net/http"

	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/endpoint_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
	routinepkg "github.com/langgenius/dify-plugin-daemon/pkg/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/stream"
)

func InvokeEndpoint(
	session *session_manager.Session,
	request *requests.RequestInvokeEndpoint,
) (
	int, *http.Header, *stream.Stream[[]byte], error,
) {
	resp, err := GenericInvokePlugin[requests.RequestInvokeEndpoint, endpoint_entities.EndpointResponseChunk](
		session,
		request,
		128,
	)

	if err != nil {
		return http.StatusInternalServerError, nil, nil, err
	}

	statusCode := http.StatusContinue
	headers := &http.Header{}
	response := stream.NewStream[[]byte](128)
	response.OnClose(func() {
		// add close callback, ensure resources are released
		resp.Close()
	})

	for resp.Next() {
		result, err := resp.Read()
		if err != nil {
			response.Close()
			return http.StatusInternalServerError, nil, nil, err
		}

		if result.Status != nil {
			statusCode = int(*result.Status)
		}

		if result.Headers != nil {
			for k, v := range result.Headers {
				headers.Add(k, v)
			}
		}

		if result.Result != nil {
			dehexed, err := hex.DecodeString(*result.Result)
			if err != nil {
				response.Close()
				return http.StatusInternalServerError, nil, nil, err
			}

			response.Write(dehexed)
			routine.Submit(routinepkg.Labels{
				routinepkg.RoutineLabelKeyModule:      "plugin_daemon",
				routinepkg.RoutineLabelKeyMethod:      "InvokeEndpoint",
				routinepkg.RoutineLabelRuntimeKeyType: "body_write",
			}, func() {
				defer response.Close()
				for resp.Next() {
					chunk, err := resp.Read()
					if err != nil {
						response.WriteError(err)
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

	if resp.IsClosed() {
		response.Close()
	}

	return statusCode, headers, response, nil
}
