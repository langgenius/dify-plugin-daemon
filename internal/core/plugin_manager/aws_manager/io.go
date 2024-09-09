package aws_manager

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

func (r *AWSPluginRuntime) Listen(session_id string) *entities.Broadcast[plugin_entities.SessionMessage] {
	l := entities.NewBroadcast[plugin_entities.SessionMessage]()
	// store the listener
	r.listeners.Store(session_id, l)
	return l
}

// For AWS Lambda, write is equivalent to http request, it's not a normal stream like stdio and tcp
func (r *AWSPluginRuntime) Write(session_id string, data []byte) {
	l, ok := r.listeners.Load(session_id)
	if !ok {
		log.Error("session %s not found", session_id)
		return
	}

	url, err := url.JoinPath(r.LambdaURL, "invoke")
	if err != nil {
		r.Error(fmt.Sprintf("Error creating request: %v", err))
		return
	}

	// create a new http request
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		r.Error(fmt.Sprintf("Error creating request: %v", err))
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Dify-Plugin-Session-ID", session_id)

	routine.Submit(func() {
		// remove the session from listeners
		defer r.listeners.Delete(session_id)

		response, err := r.client.Do(req)
		if err != nil {
			r.Error(fmt.Sprintf("Error sending request to aws lambda: %v", err))
			return
		}

		// write to data stream
		scanner := bufio.NewScanner(response.Body)
		for scanner.Scan() {
			bytes := scanner.Bytes()
			if len(bytes) == 0 {
				continue
			}

			data, err := parser.UnmarshalJsonBytes[plugin_entities.SessionMessage](bytes)
			if err != nil {
				log.Error("unmarshal json failed: %s, failed to parse session message", err.Error())
				continue
			}

			l.Send(data)
		}

		l.Close()
	})
}
