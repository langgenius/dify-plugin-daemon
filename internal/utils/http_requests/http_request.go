package http_requests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

func buildHttpRequest(method string, url string, options ...HttpOptions) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	for _, option := range options {
		switch option.Type {
		case "write_timeout":
			timeout := time.Second * time.Duration(option.Value.(int64))
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			time.AfterFunc(timeout, cancel) // release resources associated with context asynchronously
			req = req.WithContext(ctx)
		case "header":
			for k, v := range option.Value.(map[string]string) {
				req.Header.Set(k, v)
			}
		case "params":
			q := req.URL.Query()
			for k, v := range option.Value.(map[string]string) {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()
		case "payload":
			q := req.URL.Query()
			for k, v := range option.Value.(map[string]string) {
				q.Add(k, v)
			}
			req.Body = io.NopCloser(strings.NewReader(q.Encode()))
		case "payloadMultipart":
			buffer := new(bytes.Buffer)
			writer := multipart.NewWriter(buffer)

			files := option.Value.(map[string]any)["files"].(map[string]io.Reader)
			for filename, reader := range files {
				part, err := writer.CreateFormFile(filename, filename)
				if err != nil {
					writer.Close()
					return nil, err
				}
				_, err = io.Copy(part, reader)
				if err != nil {
					writer.Close()
					return nil, err
				}
			}

			payload := option.Value.(map[string]any)["payload"].(map[string]string)
			for k, v := range payload {
				if err := writer.WriteField(k, v); err != nil {
					writer.Close()
					return nil, err
				}
			}

			if err := writer.Close(); err != nil {
				return nil, err
			}

			req.Body = io.NopCloser(buffer)
			req.Header.Set("Content-Type", writer.FormDataContentType())
		case "payloadText":
			req.Body = io.NopCloser(strings.NewReader(option.Value.(string)))
			req.Header.Set("Content-Type", "text/plain")
		case "payloadJson":
			jsonStr, err := json.Marshal(option.Value)
			if err != nil {
				return nil, err
			}
			req.Body = io.NopCloser(bytes.NewBuffer(jsonStr))
			// set application/json content type
			req.Header.Set("Content-Type", "application/json")
		case "directReferer":
			req.Header.Set("Referer", url)
		}
	}

	return req, nil
}

func Request(client *http.Client, url string, method string, options ...HttpOptions) (*http.Response, error) {
	req, err := buildHttpRequest(method, url, options...)

	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
