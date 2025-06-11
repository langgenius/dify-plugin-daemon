package http_requests

import (
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
		if option.Type == HttpOptionTypeReadTimeout {
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
	usingLengthPrefixed := false
	for _, option := range options {
		if option.Type == HttpOptionTypeReadTimeout {
			readTimeout = option.Value.(int64)
		} else if option.Type == HttpOptionTypeRaiseErrorWhenStreamDataNotMatch {
			raiseErrorWhenStreamDataNotMatch = option.Value.(bool)
		} else if option.Type == HttpOptionTypeUsingLengthPrefixed {
			usingLengthPrefixed = option.Value.(bool)
		}
	}
	time.AfterFunc(time.Millisecond*time.Duration(readTimeout), func() {
		// close the response body if timeout
		resp.Body.Close()
	})

	// Common data processor function to reduce code duplication
	processData := func(data []byte) error {
		// unmarshal
		t, err := parser.UnmarshalJsonBytes[T](data)
		if err != nil {
			if raiseErrorWhenStreamDataNotMatch {
				return err
			} else {
				log.Warn("stream data not match for %s, got %s", url, string(data))
				return nil
			}
		}

		ch.Write(t)
		return nil
	}

	routine.Submit(map[string]string{
		"module":   "http_requests",
		"function": "RequestAndParseStream",
	}, func() {
		defer resp.Body.Close()

		var err error
		if usingLengthPrefixed {
			// at most 30MB a single chunk
			err = parser.LengthPrefixedChunking(resp.Body, 0x0f, 1024*1024*30, processData)
		} else {
			err = parser.LineBasedChunking(resp.Body, 1024*1024*30, func(data []byte) error {
				if len(data) == 0 {
					return nil
				}

				if bytes.HasPrefix(data, []byte("data:")) {
					// split
					data = data[5:]
				}

				if bytes.HasPrefix(data, []byte("event:")) {
					// TODO: handle event
					return nil
				}

				// trim space
				data = bytes.TrimSpace(data)

				return processData(data)
			})
		}

		if err != nil {
			ch.WriteError(err)
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
