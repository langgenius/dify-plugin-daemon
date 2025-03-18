package serverless_runtime

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
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

		// write to data stream
		scanner := bufio.NewScanner(response.Body)
		defer response.Body.Close()

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
	})
}
