package serverless_runtime

import (
	"bufio"
	"bytes"
	"encoding/json"
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

const serverlessResponsePreviewLimit = 4 * 1024

type serverlessErrorResponse struct {
	ErrorType          string `json:"errorType"`
	ErrorTypeSnakeCase string `json:"error_type"`
	RequestID          string `json:"requestId"`
	RequestIDSnakeCase string `json:"request_id"`
}

func (e serverlessErrorResponse) errorType() string {
	if e.ErrorType != "" {
		return e.ErrorType
	}
	return e.ErrorTypeSnakeCase
}

func (e serverlessErrorResponse) requestID() string {
	if e.RequestID != "" {
		return e.RequestID
	}
	return e.RequestIDSnakeCase
}

type serverlessResponsePreview struct {
	data      []byte
	bytesRead int
}

func (p *serverlessResponsePreview) Write(data []byte) (int, error) {
	written := len(data)
	p.bytesRead += written
	remaining := serverlessResponsePreviewLimit - len(p.data)
	if remaining > 0 {
		if len(data) > remaining {
			data = data[:remaining]
		}
		p.data = append(p.data, data...)
	}
	return written, nil
}

func (p *serverlessResponsePreview) bytes() []byte {
	return p.data
}

func (p *serverlessResponsePreview) string() string {
	return string(bytes.TrimSpace(p.data))
}

func (p *serverlessResponsePreview) truncated() bool {
	return p.bytesRead > len(p.data)
}

func readServerlessResponsePreview(reader io.Reader) (serverlessResponsePreview, error) {
	preview := serverlessResponsePreview{}
	_, err := io.Copy(&preview, io.LimitReader(reader, serverlessResponsePreviewLimit+1))
	return preview, err
}

func parseServerlessErrorResponse(data []byte) serverlessErrorResponse {
	var errorResponse serverlessErrorResponse
	_ = json.Unmarshal(bytes.TrimSpace(data), &errorResponse)
	return errorResponse
}

func serverlessRuntimeErrorDetails(
	response *http.Response,
	responsePreview []byte,
) (errorType string, requestID string) {
	errorType = response.Header.Get("x-amzn-ErrorType")
	requestID = response.Header.Get("x-amzn-RequestId")
	errorResponse := parseServerlessErrorResponse(responsePreview)
	if errorType == "" {
		errorType = errorResponse.errorType()
	}
	if requestID == "" {
		requestID = errorResponse.requestID()
	}
	return errorType, requestID
}

func buildServerlessRuntimeError(
	response *http.Response,
	responsePreview []byte,
	fallbackReason string,
) plugin_entities.ErrorResponse {
	errorType, requestID := serverlessRuntimeErrorDetails(response, responsePreview)
	if errorType == "" {
		errorType = fallbackReason
	}
	if errorType == "" {
		errorType = fmt.Sprintf("HTTP %d", response.StatusCode)
	}

	args := map[string]any{
		"status_code": response.StatusCode,
	}
	if requestID != "" {
		args["request_id"] = requestID
	}

	return plugin_entities.ErrorResponse{
		ErrorType: "PluginRuntimeError",
		Message:   fmt.Sprintf("Plugin runtime request failed: %s", errorType),
		Args:      args,
	}
}

func logServerlessResponseFailure(
	message string,
	sessionID string,
	action access_types.PluginAccessAction,
	payloadSize int,
	response *http.Response,
	preview *serverlessResponsePreview,
	previewErr error,
) {
	errorType, requestID := serverlessRuntimeErrorDetails(response, preview.bytes())
	args := []any{
		"session_id", sessionID,
		"action", action,
		"payload_size_bytes", payloadSize,
		"status_code", response.StatusCode,
		"status", response.Status,
		"http_success", response.StatusCode >= 200 && response.StatusCode < 300,
		"lambda_request_id", requestID,
		"lambda_error_type", errorType,
		"content_type", response.Header.Get("Content-Type"),
		"content_length", response.ContentLength,
		"response_bytes_read", preview.bytesRead,
		"response_preview", preview.string(),
		"response_preview_truncated", preview.truncated(),
	}
	if previewErr != nil {
		args = append(args, "response_preview_error", previewErr)
	}
	log.Error(message, args...)
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
			log.Warn(
				"serverless runtime HTTP request attempt failed",
				"session_id", sessionId,
				"attempt", attempt+1,
				"max_attempts", maxRetries,
				"payload_size_bytes", len(data),
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
			preview := serverlessResponsePreview{}
			var previewErr error
			if response.Body != nil {
				preview, previewErr = readServerlessResponsePreview(response.Body)
				response.Body.Close()
			}
			errorType, requestID := serverlessRuntimeErrorDetails(response, preview.bytes())
			log.Warn(
				"serverless runtime HTTP response will be retried",
				"session_id", sessionId,
				"attempt", attempt+1,
				"max_attempts", maxRetries,
				"payload_size_bytes", len(data),
				"status_code", response.StatusCode,
				"status", response.Status,
				"lambda_request_id", requestID,
				"lambda_error_type", errorType,
				"content_type", response.Header.Get("Content-Type"),
				"content_length", response.ContentLength,
				"response_bytes_read", preview.bytesRead,
				"response_preview", preview.string(),
				"response_preview_truncated", preview.truncated(),
				"response_preview_error", previewErr,
			)
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
		sendEnd := true
		sendError := func(errorResponse plugin_entities.ErrorResponse) {
			if !sendEnd {
				return
			}
			sendEnd = false
			l.Send(plugin_entities.SessionMessage{
				Type: plugin_entities.SESSION_MESSAGE_TYPE_ERROR,
				Data: parser.MarshalJsonBytes(errorResponse),
			})
		}
		defer func() {
			if sendEnd {
				l.Send(plugin_entities.SessionMessage{
					Type: plugin_entities.SESSION_MESSAGE_TYPE_END,
					Data: []byte(""),
				})
			}
			l.Close()
			r.listeners.Delete(sessionId)
		}()

		url += "?action=" + string(action)
		response, err := r.invokeServerlessWithRetry(url, sessionId, data)
		if err != nil {
			log.Error(
				"serverless runtime invocation failed before receiving a response",
				"session_id", sessionId,
				"action", action,
				"payload_size_bytes", len(data),
				"error", err,
			)
			sendError(plugin_entities.ErrorResponse{
				ErrorType: "PluginDaemonInnerError",
				Message:   fmt.Sprintf("Error sending request to serverless: %v", err),
			})
			return
		}

		defer response.Body.Close()
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			preview, previewErr := readServerlessResponsePreview(response.Body)
			logServerlessResponseFailure(
				"serverless runtime returned non-success HTTP response",
				sessionId,
				action,
				len(data),
				response,
				&preview,
				previewErr,
			)
			sendError(buildServerlessRuntimeError(response, preview.bytes(), response.Status))
			return
		}

		if response.Header.Get("x-amzn-ErrorType") != "" {
			preview, previewErr := readServerlessResponsePreview(response.Body)
			logServerlessResponseFailure(
				"serverless runtime returned Lambda error headers with successful HTTP status",
				sessionId,
				action,
				len(data),
				response,
				&preview,
				previewErr,
			)
			sendError(buildServerlessRuntimeError(response, preview.bytes(), "Lambda runtime error"))
			return
		}

		preview := serverlessResponsePreview{}
		scanner := bufio.NewScanner(io.TeeReader(response.Body, &preview))

		scanner.Buffer(make([]byte, r.RuntimeBufferSize), r.RuntimeMaxBufferSize)

		sessionAlive := true
		hasSessionEvent := false
		for scanner.Scan() && sessionAlive {
			line := scanner.Bytes()

			if len(line) == 0 {
				continue
			}

			lambdaError := parseServerlessErrorResponse(line)
			if lambdaError.errorType() != "" {
				logServerlessResponseFailure(
					"serverless runtime returned Lambda error payload with successful HTTP status",
					sessionId,
					action,
					len(data),
					response,
					&preview,
					nil,
				)
				sendError(buildServerlessRuntimeError(response, line, "Lambda runtime error"))
				break
			}

			plugin_entities.ParsePluginUniversalEvent(
				line,
				response.Status,
				func(session_id string, sessionData []byte) {
					sessionMessage, err := parser.UnmarshalJsonBytes[plugin_entities.SessionMessage](sessionData)
					if err != nil {
						logServerlessResponseFailure(
							"serverless runtime returned an invalid session message",
							sessionId,
							action,
							len(data),
							response,
							&preview,
							nil,
						)
						sendError(plugin_entities.ErrorResponse{
							ErrorType: "PluginDaemonInnerError",
							Message:   fmt.Sprintf("failed to parse session message %s, err: %v", line, err),
						})
						sessionAlive = false
						return
					}
					hasSessionEvent = true
					l.Send(sessionMessage)
				},
				func() {},
				func(err string) {
					logServerlessResponseFailure(
						"serverless runtime returned an invalid plugin event",
						sessionId,
						action,
						len(data),
						response,
						&preview,
						nil,
					)
					sendError(plugin_entities.ErrorResponse{
						ErrorType: "PluginDaemonInnerError",
						Message:   fmt.Sprintf("encountered an error: %v", err),
					})
					sessionAlive = false
				},
				func(plugin_entities.PluginLogEvent) {},
			)
		}

		if err := scanner.Err(); err != nil {
			logServerlessResponseFailure(
				"serverless runtime response body could not be read",
				sessionId,
				action,
				len(data),
				response,
				&preview,
				err,
			)
			sendError(plugin_entities.ErrorResponse{
				ErrorType: "PluginDaemonInnerError",
				Message:   fmt.Sprintf("failed to read response body: %v", err),
			})
			return
		}

		if !hasSessionEvent && sendEnd {
			logServerlessResponseFailure(
				"serverless runtime returned no valid session response",
				sessionId,
				action,
				len(data),
				response,
				&preview,
				nil,
			)
			sendError(buildServerlessRuntimeError(
				response,
				preview.bytes(),
				"no valid session response",
			))
		}
	})

	return nil
}
