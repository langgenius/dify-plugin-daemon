package aws_manager

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

// consume data from data stream
func (r *AWSPluginRuntime) consume() {
	for {
		select {
		case data := <-r.data_stream:
			fmt.Println(data)
		}
	}
}

func (r *AWSPluginRuntime) Listen(session_id string) *entities.BytesIOListener {
	l := entities.NewIOListener[[]byte]()
	l.OnClose(func() {
		// close the pipe writer
		writer, exists := r.session_pool.Load(session_id)
		if exists {
			writer.Close()
		}
	})
	return l
}

func (r *AWSPluginRuntime) Write(session_id string, data []byte) {
	// check if session exists
	var pw *io.PipeWriter
	var exists bool

	if pw, exists = r.session_pool.Load(session_id); !exists {
		url, err := url.JoinPath(r.lambda_url, "invoke")
		if err != nil {
			r.Error(fmt.Sprintf("Error creating request: %v", err))
			return
		}

		// create a new http request here
		npr, npw := io.Pipe()
		r.session_pool.Store(session_id, npw)
		pw = npw

		// create a new http request
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, "POST", url, npr)
		if err != nil {
			r.Error(fmt.Sprintf("Error creating request: %v", err))
			return
		}

		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("Accept", "application/octet-stream")

		routine.Submit(func() {
			response, err := http.DefaultClient.Do(req)
			if err != nil {
				r.Error(fmt.Sprintf("Error sending request to aws lambda: %v", err))
				return
			}

			// write to data stream
			for {
				buf := make([]byte, 1024)
				n, err := response.Body.Read(buf)
				if err != nil {
					if err == io.EOF {
						break
					} else {
						r.Error(fmt.Sprintf("Error reading response from aws lambda: %v", err))
						break
					}
				}
				// write to data stream
				select {
				case r.data_stream <- buf[:n]:
				default:
					r.Error("Data stream is full")
				}
			}

			// remove the session from the pool
			r.session_pool.Delete(session_id)
		})
	}

	if pw != nil {
		if _, err := pw.Write(data); err != nil {
			r.Error(fmt.Sprintf("Error writing to pipe writer: %v", err))
		}
	}
}
