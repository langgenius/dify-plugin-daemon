package serverless_runtime

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

const serverlessResponsePreviewLimit = 4 * 1024

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

func (r *AWSPluginRuntime) Listen(sessionId string) *entities.Broadcast[plugin_entities.SessionMessage] {
	l := entities.NewBroadcast[plugin_entities.SessionMessage]()
	// store the listener
	r.listeners.Store(sessionId, l)
	return l
}

// For AWS Lambda, write is equivalent to http request, it's not a normal stream like stdio and tcp
func (r *AWSPluginRuntime) Write(sessionId string, action access_types.PluginAccessAction, data []byte) {
	l, ok := r.listeners.Load(sessionId)
	if !ok {
		log.Error("session %s not found", sessionId)
		return
	}

	url, err := url.JoinPath(r.LambdaURL, "invoke")
	if err != nil {
		l.Send(plugin_entities.SessionMessage{
			Type: plugin_entities.SESSION_MESSAGE_TYPE_ERROR,
			Data: parser.MarshalJsonBytes(plugin_entities.ErrorResponse{
				ErrorType: "PluginDaemonInnerError",
				Message:   fmt.Sprintf("Error creating request: %v", err),
			}),
		})
		l.Close()
		r.Error(fmt.Sprintf("Error creating request: %v", err))
		return
	}

	url += "?action=" + string(action)

	connectTime := 240 * time.Second
	payloadSizeBytes := len(data)

	// create a new http request
	ctx, cancel := context.WithTimeout(context.Background(), connectTime)
	time.AfterFunc(connectTime, cancel)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		r.Error(fmt.Sprintf("Error creating request: %v", err))
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Dify-Plugin-Session-ID", sessionId)

	log.Info(
		"sending request to aws lambda, session_id=%s action=%s lambda_url=%s payload_size_bytes=%d",
		sessionId,
		action,
		r.LambdaURL,
		payloadSizeBytes,
	)

	routine.Submit(map[string]string{
		"module":     "serverless_runtime",
		"function":   "Write",
		"session_id": sessionId,
		"lambda_url": r.LambdaURL,
	}, func() {
		// remove the session from listeners
		defer r.listeners.Delete(sessionId)
		defer l.Close()
		defer l.Send(plugin_entities.SessionMessage{
			Type: plugin_entities.SESSION_MESSAGE_TYPE_END,
			Data: []byte(""),
		})

		response, err := r.client.Do(req)
		if err != nil {
			l.Send(plugin_entities.SessionMessage{
				Type: plugin_entities.SESSION_MESSAGE_TYPE_ERROR,
				Data: parser.MarshalJsonBytes(plugin_entities.ErrorResponse{
					ErrorType: "PluginDaemonInnerError",
					Message:   fmt.Sprintf("Error sending request to aws lambda: %v", err),
				}),
			})
			r.Error(fmt.Sprintf("Error sending request to aws lambda: %v", err))
			return
		}

		defer response.Body.Close()

		// write to data stream
		responsePreview := &limitedBuffer{limit: serverlessResponsePreviewLimit}
		scanner := bufio.NewScanner(io.TeeReader(response.Body, responsePreview))

		// TODO: set a reasonable buffer size or use a reader, this is a temporary solution
		scanner.Buffer(make([]byte, 1024), 5*1024*1024)

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
				func(message string) {},
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
			log.Warn(
				"aws lambda returned non-success status, session_id=%s action=%s lambda_url=%s payload_size_bytes=%d status=%s status_code=%d headers=%v response_preview=%q",
				sessionId,
				action,
				r.LambdaURL,
				payloadSizeBytes,
				response.Status,
				response.StatusCode,
				serverlessResponseHeaders(response),
				responsePreview.String(),
			)
		}
	})
}
