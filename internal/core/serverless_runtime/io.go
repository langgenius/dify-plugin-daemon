package serverless_runtime

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	routinepkg "github.com/langgenius/dify-plugin-daemon/pkg/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/http_requests"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
)

func (r *ServerlessPluginRuntime) Listen(sessionId string) (
	*entities.Broadcast[plugin_entities.SessionMessage],
	error,
) {
	l := entities.NewCallbackHandler[plugin_entities.SessionMessage]()
	// store the listener
	r.listeners.Store(sessionId, l)
	return l, nil
}

// shouldRetryStatusCode checks if the HTTP status code warrants a retry
// Only 502 (Bad Gateway) errors are retried as they typically indicate temporary gateway issues
//
// To some AWS Lambda gateway errors, 502 randomly happens, and it's usually transient.
// Thus we implement a retry mechanism for 502 errors.
func shouldRetryStatusCode(statusCode int) bool {
	return statusCode == 502
}

// invokeServerlessWithRetry invokes the serverless endpoint with retry logic
// It will retry up to MaxRetryTimes attempts on 502 errors with exponential backoff
func (r *ServerlessPluginRuntime) invokeServerlessWithRetry(
	url string,
	sessionId string,
	data []byte,
) (*http.Response, error) {
	var lastErr error
	var response *http.Response

	maxRetries := r.MaxRetryTimes
	if maxRetries <= 0 {
		maxRetries = 1
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Apply exponential backoff for retry attempts (500ms, 1000ms, 2000ms, ...)
		if attempt > 0 {
			backoffDuration := time.Duration(500*(1<<uint(attempt-1))) * time.Millisecond
			time.Sleep(backoffDuration)
		}

		// Make HTTP request to serverless endpoint
		response, err := http_requests.Request(
			r.Client, url, "POST",
			http_requests.HttpHeader(map[string]string{
				"Content-Type":           "application/json",
				"Accept":                 "text/event-stream",
				"Dify-Plugin-Session-ID": sessionId,
			}),
			http_requests.HttpPayloadReader(io.NopCloser(bytes.NewReader(data))),
			http_requests.HttpReadTimeout(int64(r.PluginMaxExecutionTimeout*1000)),
		)

		if err != nil {
			lastErr = fmt.Errorf("attempt %d/%d failed: %w", attempt+1, maxRetries, err)
			continue
		}

		statusCode := response.StatusCode
		// Success - return immediately
		if statusCode >= 200 && statusCode < 300 {
			return response, nil
		}

		// Check if status code should trigger a retry (502 Bad Gateway only)
		if shouldRetryStatusCode(statusCode) {
			if response.Body != nil {
				response.Body.Close()
			}
			lastErr = fmt.Errorf("attempt %d/%d failed with status code: %d", attempt+1, maxRetries, statusCode)
			continue
		}

		// Non-retryable error - return immediately
		return response, nil
	}

	if lastErr != nil {
		return nil, fmt.Errorf("all %d attempts failed, last error: %w", maxRetries, lastErr)
	}

	return response, nil
}

// For Serverless, write is equivalent to http request, it's not a normal stream like stdio and tcp
func (r *ServerlessPluginRuntime) Write(
	sessionId string,
	action access_types.PluginAccessAction,
	data []byte,
) error {
	l, ok := r.listeners.Load(sessionId)
	if !ok {
		return errors.New("session not found")
	}

	url, err := url.JoinPath(r.LambdaURL, "invoke")
	if err != nil {
		return errors.Join(err, errors.New("failed to join lambda url"))
	}

	routine.Submit(routinepkg.Labels{
		routinepkg.RoutineLabelKeyModule:    "serverless_runtime",
		routinepkg.RoutineLabelKeyMethod:    "Write",
		routinepkg.RoutineLabelKeySessionID: sessionId,
		routinepkg.RoutineLabelKeyLambdaURL: r.LambdaURL,
	}, func() {
		defer r.listeners.Delete(sessionId)
		defer l.Close()
		defer l.Send(plugin_entities.SessionMessage{
			Type: plugin_entities.SESSION_MESSAGE_TYPE_END,
			Data: []byte(""),
		})

		url += "?action=" + string(action)
		response, err := r.invokeServerlessWithRetry(url, sessionId, data)
		if err != nil {
			l.Send(plugin_entities.SessionMessage{
				Type: plugin_entities.SESSION_MESSAGE_TYPE_ERROR,
				Data: parser.MarshalJsonBytes(plugin_entities.ErrorResponse{
					ErrorType: "PluginDaemonInnerError",
					Message:   fmt.Sprintf("Error sending request to serverless: %v", err),
				}),
			})
			return
		}

		scanner := bufio.NewScanner(response.Body)
		defer response.Body.Close()

		scanner.Buffer(make([]byte, r.RuntimeBufferSize), r.RuntimeMaxBufferSize)

		sessionAlive := true
		for scanner.Scan() && sessionAlive {
			bytes := scanner.Bytes()

			if len(bytes) == 0 {
				continue
			}

			plugin_entities.ParsePluginUniversalEvent(
				bytes,
				response.Status,
				func(session_id string, data []byte) {
					sessionMessage, err := parser.UnmarshalJsonBytes[plugin_entities.SessionMessage](data)
					if err != nil {
						l.Send(plugin_entities.SessionMessage{
							Type: plugin_entities.SESSION_MESSAGE_TYPE_ERROR,
							Data: parser.MarshalJsonBytes(plugin_entities.ErrorResponse{
								ErrorType: "PluginDaemonInnerError",
								Message:   fmt.Sprintf("failed to parse session message %s, err: %v", bytes, err),
							}),
						})
						sessionAlive = false
					}
					l.Send(sessionMessage)
				},
				func() {},
				func(err string) {
					l.Send(plugin_entities.SessionMessage{
						Type: plugin_entities.SESSION_MESSAGE_TYPE_ERROR,
						Data: parser.MarshalJsonBytes(plugin_entities.ErrorResponse{
							ErrorType: "PluginDaemonInnerError",
							Message:   fmt.Sprintf("encountered an error: %v", err),
						}),
					})
				},
				func(plugin_entities.PluginLogEvent) {},
			)
		}

		if err := scanner.Err(); err != nil {
			l.Send(plugin_entities.SessionMessage{
				Type: plugin_entities.SESSION_MESSAGE_TYPE_ERROR,
				Data: parser.MarshalJsonBytes(plugin_entities.ErrorResponse{
					ErrorType: "PluginDaemonInnerError",
					Message:   fmt.Sprintf("failed to read response body: %v", err),
				}),
			})
		}
	})

	return nil
}
