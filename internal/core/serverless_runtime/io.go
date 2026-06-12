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
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
)

const serverlessErrorResponsePreviewLimit = 4 * 1024

type limitedBuffer struct {
	buf   bytes.Buffer
	limit int
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	originalLen := len(p)
	if b.limit <= 0 || b.buf.Len() >= b.limit {
		return originalLen, nil
	}

	remaining := b.limit - b.buf.Len()
	if len(p) > remaining {
		p = p[:remaining]
	}

	_, err := b.buf.Write(p)
	return originalLen, err
}

func (b *limitedBuffer) String() string {
	return b.buf.String()
}

func serverlessResponseHeaders(response *http.Response) map[string]string {
	headers := map[string]string{}
	for _, key := range []string{
		"x-amzn-RequestId",
		"x-amzn-ErrorType",
		"x-amz-function-error",
		"content-type",
		"content-length",
	} {
		if value := response.Header.Get(key); value != "" {
			headers[key] = value
		}
	}
	return headers
}

func readServerlessErrorPreview(response *http.Response) (string, error) {
	if response.Body == nil {
		return "", nil
	}

	preview, err := io.ReadAll(io.LimitReader(response.Body, serverlessErrorResponsePreviewLimit))
	if err != nil {
		return "", err
	}
	return string(preview), nil
}

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
// Backoff duration is capped at 30 seconds to prevent unreasonable wait times
func (r *ServerlessPluginRuntime) invokeServerlessWithRetry(
	url string,
	sessionId string,
	data []byte,
) (*http.Response, error) {
	const maxBackoffDuration = 30 * time.Second

	var lastErr error

	maxRetries := r.MaxRetryTimes
	if maxRetries <= 0 {
		maxRetries = 1
	}

	payloadSizeBytes := len(data)

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Apply exponential backoff for retry attempts (500ms, 1000ms, 2000ms, ...)
		// Capped at 30 seconds to prevent unreasonable wait times
		if attempt > 0 {
			backoffDuration := time.Duration(500*(1<<uint(attempt-1))) * time.Millisecond
			if backoffDuration > maxBackoffDuration {
				backoffDuration = maxBackoffDuration
			}
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
			log.Warn("serverless request failed",
				"session_id", sessionId,
				"attempt", attempt+1,
				"max_retries", maxRetries,
				"url", url,
				"payload_size_bytes", payloadSizeBytes,
				"error", err,
			)
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
			responsePreview, readErr := readServerlessErrorPreview(response)
			log.Warn("serverless request returned retryable status",
				"session_id", sessionId,
				"attempt", attempt+1,
				"max_retries", maxRetries,
				"url", url,
				"payload_size_bytes", payloadSizeBytes,
				"status", response.Status,
				"status_code", statusCode,
				"headers", serverlessResponseHeaders(response),
				"response_preview", responsePreview,
				"response_preview_read_error", readErr,
			)
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

	return nil, fmt.Errorf("all %d attempts failed with unknown error", maxRetries)
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

		responsePreview := &limitedBuffer{limit: serverlessErrorResponsePreviewLimit}
		scanner := bufio.NewScanner(io.TeeReader(response.Body, responsePreview))
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

		if response.StatusCode < 200 || response.StatusCode >= 300 {
			log.Warn("serverless request returned non-success status",
				"session_id", sessionId,
				"action", action,
				"lambda_url", r.LambdaURL,
				"payload_size_bytes", len(data),
				"status", response.Status,
				"status_code", response.StatusCode,
				"headers", serverlessResponseHeaders(response),
				"response_preview", responsePreview.String(),
			)
		}
	})

	return nil
}
