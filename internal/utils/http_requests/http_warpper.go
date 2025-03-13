package http_requests

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

func parseJsonBody(resp *http.Response, ret interface{}) error {
	defer resp.Body.Close()
	jsonDecoder := json.NewDecoder(resp.Body)
	return jsonDecoder.Decode(ret)
}

func RequestAndParse[T any](client *http.Client, url string, method string, options ...HttpOptions) (*T, error) {
	var ret T

	// check if ret is a map, if so, create a new map
	if _, ok := any(ret).(map[string]any); ok {
		ret = *new(T)
	}

	resp, err := Request(client, url, method, options...)
	if err != nil {
		return nil, err
	}

	// get read timeout
	readTimeout := int64(60000)
	for _, option := range options {
		if option.Type == "read_timeout" {
			readTimeout = option.Value.(int64)
			break
		}
	}
	time.AfterFunc(time.Millisecond*time.Duration(readTimeout), func() {
		// close the response body if timeout
		resp.Body.Close()
	})

	err = parseJsonBody(resp, &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func GetAndParse[T any](client *http.Client, url string, options ...HttpOptions) (*T, error) {
	return RequestAndParse[T](client, url, "GET", options...)
}

func PostAndParse[T any](client *http.Client, url string, options ...HttpOptions) (*T, error) {
	return RequestAndParse[T](client, url, "POST", options...)
}

func PutAndParse[T any](client *http.Client, url string, options ...HttpOptions) (*T, error) {
	return RequestAndParse[T](client, url, "PUT", options...)
}

func DeleteAndParse[T any](client *http.Client, url string, options ...HttpOptions) (*T, error) {
	return RequestAndParse[T](client, url, "DELETE", options...)
}

func PatchAndParse[T any](client *http.Client, url string, options ...HttpOptions) (*T, error) {
	return RequestAndParse[T](client, url, "PATCH", options...)
}

func RequestAndParseStream[T any](client *http.Client, url string, method string, options ...HttpOptions) (*stream.Stream[T], error) {
	resp, err := Request(client, url, method, options...)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		errorText, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status code: %d and respond with: %s", resp.StatusCode, errorText)
	}

	ch := stream.NewStream[T](1024)

	// get read timeout
	readTimeout := int64(60000)
	raiseErrorWhenStreamDataNotMatch := false
	for _, option := range options {
		if option.Type == "read_timeout" {
			readTimeout = option.Value.(int64)
			break
		} else if option.Type == "raiseErrorWhenStreamDataNotMatch" {
			raiseErrorWhenStreamDataNotMatch = option.Value.(bool)
		}
	}
	time.AfterFunc(time.Millisecond*time.Duration(readTimeout), func() {
		// close the response body if timeout
		resp.Body.Close()
	})

	routine.Submit(map[string]string{
		"module":   "http_requests",
		"function": "RequestAndParseStream",
	}, func() {
		scanner := bufio.NewScanner(resp.Body)
		defer resp.Body.Close()

		for scanner.Scan() {
			data := scanner.Bytes()
			if len(data) == 0 {
				continue
			}

			if bytes.HasPrefix(data, []byte("data:")) {
				// split
				data = data[5:]
			}

			if bytes.HasPrefix(data, []byte("event:")) {
				// TODO: handle event
				continue
			}

			// trim space
			data = bytes.TrimSpace(data)

			// unmarshal
			t, err := parser.UnmarshalJsonBytes[T](data)
			if err != nil {
				if raiseErrorWhenStreamDataNotMatch {
					ch.WriteError(err)
					break
				} else {
					log.Warn("stream data not match for %s, got %s", url, string(data))
				}
				continue
			}

			ch.Write(t)
		}

		ch.Close()
	})

	return ch, nil
}

func GetAndParseStream[T any](client *http.Client, url string, options ...HttpOptions) (*stream.Stream[T], error) {
	return RequestAndParseStream[T](client, url, "GET", options...)
}

func PostAndParseStream[T any](client *http.Client, url string, options ...HttpOptions) (*stream.Stream[T], error) {
	return RequestAndParseStream[T](client, url, "POST", options...)
}

func PutAndParseStream[T any](client *http.Client, url string, options ...HttpOptions) (*stream.Stream[T], error) {
	return RequestAndParseStream[T](client, url, "PUT", options...)
}
