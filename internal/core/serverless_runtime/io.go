package serverless_runtime

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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

const serverlessErrorResponseLimit = 4 * 1024

// ServerlessPayloadTooLargeError reports request size metadata without exposing payload contents.
type ServerlessPayloadTooLargeError struct {
	Action          access_types.PluginAccessAction
	PayloadBytes    int
	MaxRequestBytes int
	ItemIndex       *int
}

func (e *ServerlessPayloadTooLargeError) Error() string {
	message := fmt.Sprintf(
		"ServerlessPayloadTooLarge: action=%s payload_bytes=%d max_request_bytes=%d",
		e.Action,
		e.PayloadBytes,
		e.MaxRequestBytes,
	)
	if e.ItemIndex != nil {
		message += fmt.Sprintf(" item_index=%d", *e.ItemIndex)
	}
	return message
}

type serverlessErrorResponse struct {
	ErrorType    string `json:"errorType"`
	ErrorMessage string `json:"errorMessage"`
	RequestID    string `json:"requestId"`
}

func parseServerlessErrorResponse(data []byte) serverlessErrorResponse {
	var errorResponse serverlessErrorResponse
	_ = json.Unmarshal(data, &errorResponse)
	return errorResponse
}

func readServerlessResponseBody(reader io.Reader) ([]byte, error) {
	return io.ReadAll(io.LimitReader(reader, serverlessErrorResponseLimit))
}

func serverlessRuntimeErrorDetails(
	response *http.Response,
	responseBody []byte,
) serverlessErrorResponse {
	details := parseServerlessErrorResponse(responseBody)
	if errorType := response.Header.Get("x-amzn-ErrorType"); errorType != "" {
		details.ErrorType = errorType
	}
	if requestID := response.Header.Get("x-amzn-RequestId"); requestID != "" {
		details.RequestID = requestID
	}
	return details
}

func buildServerlessRuntimeError(
	response *http.Response,
	responseBody []byte,
	fallbackReason string,
) plugin_entities.ErrorResponse {
	details := serverlessRuntimeErrorDetails(response, responseBody)
	if details.ErrorType == "" {
		details.ErrorType = fallbackReason
	}
	if details.ErrorType == "" {
		details.ErrorType = fmt.Sprintf("HTTP %d", response.StatusCode)
	}

	args := map[string]any{
		"status_code": response.StatusCode,
	}
	if details.RequestID != "" {
		args["request_id"] = details.RequestID
	}
	message := fmt.Sprintf("Plugin runtime request failed: %s", details.ErrorType)
	if details.ErrorMessage != "" {
		message += ": " + details.ErrorMessage
	}

	return plugin_entities.ErrorResponse{
		ErrorType: "PluginRuntimeError",
		Message:   message,
		Args:      args,
	}
}

func logServerlessResponseFailure(
	message string,
	sessionID string,
	action access_types.PluginAccessAction,
	payloadSize int,
	response *http.Response,
	responseBody []byte,
	responseBodyErr error,
) {
	details := serverlessRuntimeErrorDetails(response, responseBody)
	args := []any{
		"session_id", sessionID,
		"action", action,
		"payload_size_bytes", payloadSize,
		"status_code", response.StatusCode,
		"lambda_request_id", details.RequestID,
		"lambda_error_type", details.ErrorType,
		"content_type", response.Header.Get("Content-Type"),
		"content_length", response.ContentLength,
		"response_body_size_bytes", len(responseBody),
	}
	if responseBodyErr != nil {
		args = append(args, "response_body_error", responseBodyErr)
	}
	log.Error(message, args...)
}

func logServerlessPluginEvent(
	sessionID string,
	action access_types.PluginAccessAction,
	lambdaName string,
	event plugin_entities.PluginLogEvent,
) {
	args := []any{
		"session_id", sessionID,
		"action", action,
		"lambda_name", lambdaName,
		"plugin_log_level", event.Level,
		"plugin_log_timestamp", event.Timestamp,
		"message", event.Message,
	}
	switch strings.ToLower(event.Level) {
	case "debug":
		log.Debug("serverless plugin log", args...)
	case "warn", "warning":
		log.Warn("serverless plugin log", args...)
	case "error", "fatal", "critical":
		log.Error("serverless plugin log", args...)
	default:
		log.Info("serverless plugin log", args...)
	}
}

func (r *ServerlessPluginRuntime) Listen(sessionId string) (
	*entities.Broadcast[plugin_entities.SessionMessage],
	error,
) {
	l := entities.NewCallbackHandler[plugin_entities.SessionMessage]()
	r.listeners.Store(sessionId, l)
	l.OnClose(func() {
		r.listeners.DeleteIf(sessionId, func(listener *entities.Broadcast[plugin_entities.SessionMessage]) bool {
			return listener == l
		})
	})
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
	ctx context.Context,
	url string,
	sessionId string,
	data []byte,
	action access_types.PluginAccessAction,
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
			timer := time.NewTimer(backoffDuration)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			case <-timer.C:
			}
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
			http_requests.HttpContext(ctx),
		)

		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			recordServerlessRuntimeFailure(action, 0, serverlessRuntimeFailureType(0))
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
		recordServerlessRuntimeFailure(action, statusCode, serverlessRuntimeFailureType(statusCode))

		// Check if status code should trigger a retry (502 Bad Gateway only)
		if shouldRetryStatusCode(statusCode) {
			if response.Body != nil {
				response.Body.Close()
			}
			log.Warn(
				"serverless runtime HTTP response will be retried",
				"session_id", sessionId,
				"attempt", attempt+1,
				"max_attempts", maxRetries,
				"payload_size_bytes", len(data),
				"status_code", response.StatusCode,
				"lambda_request_id", response.Header.Get("x-amzn-RequestId"),
				"lambda_error_type", response.Header.Get("x-amzn-ErrorType"),
				"content_type", response.Header.Get("Content-Type"),
				"content_length", response.ContentLength,
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
	return r.WriteContext(context.Background(), sessionId, action, data)
}

func (r *ServerlessPluginRuntime) WriteContext(
	ctx context.Context,
	sessionId string,
	action access_types.PluginAccessAction,
	data []byte,
) error {
	l, ok := r.listeners.Load(sessionId)
	if !ok {
		return errors.New("session not found")
	}
	deleteListener := func() {
		r.listeners.DeleteIf(sessionId, func(listener *entities.Broadcast[plugin_entities.SessionMessage]) bool {
			return listener == l
		})
	}

	url, err := url.JoinPath(r.LambdaURL, "invoke")
	if err != nil {
		deleteListener()
		return errors.Join(err, errors.New("failed to join lambda url"))
	}

	if len(data) > r.MaxRequestBytes {
		deleteListener()
		return &ServerlessPayloadTooLargeError{
			Action:          action,
			PayloadBytes:    len(data),
			MaxRequestBytes: r.MaxRequestBytes,
		}
	}

	routine.Submit(routinepkg.Labels{
		routinepkg.RoutineLabelKeyModule:    "serverless_runtime",
		routinepkg.RoutineLabelKeyMethod:    "Write",
		routinepkg.RoutineLabelKeySessionID: sessionId,
		routinepkg.RoutineLabelKeyLambdaURL: r.LambdaURL,
	}, func() {
		requestCtx := ctx
		cancel := func() {}
		if r.PluginMaxExecutionTimeout > 0 {
			requestCtx, cancel = context.WithTimeout(ctx, time.Duration(r.PluginMaxExecutionTimeout)*time.Second)
		}
		defer cancel()

		sendEnd := true
		var endData json.RawMessage
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
			deleteListener()
			if sendEnd {
				l.Send(plugin_entities.SessionMessage{
					Type: plugin_entities.SESSION_MESSAGE_TYPE_END,
					Data: endData,
				})
			}
			l.Close()
		}()

		url += "?action=" + string(action)
		response, err := r.invokeServerlessWithRetry(requestCtx, url, sessionId, data, action)
		if err != nil {
			if ctx.Err() != nil {
				sendEnd = false
				return
			}
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
		logFailure := func(message string, responseBody []byte, responseBodyErr error) {
			logServerlessResponseFailure(
				message,
				sessionId,
				action,
				len(data),
				response,
				responseBody,
				responseBodyErr,
			)
		}
		sendRuntimeError := func(message string, responseBody []byte, responseBodyErr error, fallbackReason string) {
			logFailure(message, responseBody, responseBodyErr)
			sendError(buildServerlessRuntimeError(response, responseBody, fallbackReason))
		}

		if response.StatusCode < 200 || response.StatusCode >= 300 {
			responseBody, responseBodyErr := readServerlessResponseBody(response.Body)
			sendRuntimeError(
				"serverless runtime returned non-success HTTP response",
				responseBody,
				responseBodyErr,
				response.Status,
			)
			return
		}

		if response.Header.Get("x-amzn-ErrorType") != "" {
			responseBody, responseBodyErr := readServerlessResponseBody(response.Body)
			sendRuntimeError(
				"serverless runtime returned Lambda error headers with successful HTTP status",
				responseBody,
				responseBodyErr,
				"Lambda runtime error",
			)
			return
		}

		scanner := bufio.NewScanner(response.Body)

		scanner.Buffer(make([]byte, r.RuntimeBufferSize), r.RuntimeMaxBufferSize)

		hasSessionEvent := false
		receivedEnd := false
		for sendEnd && scanner.Scan() {
			line := scanner.Bytes()

			if len(line) == 0 {
				continue
			}

			lambdaError := parseServerlessErrorResponse(line)
			if lambdaError.ErrorType != "" {
				sendRuntimeError(
					"serverless runtime returned Lambda error payload with successful HTTP status",
					line,
					nil,
					"Lambda runtime error",
				)
				break
			}

			plugin_entities.ParsePluginUniversalEvent(
				line,
				response.Status,
				func(session_id string, sessionData []byte) {
					sessionMessage, err := parser.UnmarshalJsonBytes[plugin_entities.SessionMessage](sessionData)
					if err != nil {
						logFailure(
							"serverless runtime returned an invalid session message",
							line,
							nil,
						)
						sendError(plugin_entities.ErrorResponse{
							ErrorType: "PluginDaemonInnerError",
							Message:   fmt.Sprintf("failed to parse session message %s, err: %v", line, err),
						})
						return
					}
					hasSessionEvent = true
					if sessionMessage.Type == plugin_entities.SESSION_MESSAGE_TYPE_END {
						receivedEnd = true
						endData = sessionMessage.Data
						return
					}
					l.Send(sessionMessage)
				},
				func() {},
				func(err string) {
					logFailure(
						"serverless runtime returned an invalid plugin event",
						line,
						nil,
					)
					sendError(plugin_entities.ErrorResponse{
						ErrorType: "PluginDaemonInnerError",
						Message:   fmt.Sprintf("encountered an error: %v", err),
					})
				},
				func(logEvent plugin_entities.PluginLogEvent) {
					logServerlessPluginEvent(sessionId, action, r.LambdaName, logEvent)
				},
			)
			if receivedEnd {
				break
			}
		}

		if err := scanner.Err(); err != nil {
			if ctx.Err() != nil {
				sendEnd = false
				return
			}
			logFailure(
				"serverless runtime response body could not be read",
				nil,
				err,
			)
			sendError(plugin_entities.ErrorResponse{
				ErrorType: "PluginDaemonInnerError",
				Message:   fmt.Sprintf("failed to read response body: %v", err),
			})
			return
		}

		if !hasSessionEvent && sendEnd {
			sendRuntimeError(
				"serverless runtime returned no valid session response",
				nil,
				nil,
				"no valid session response",
			)
		}
	})

	return nil
}
