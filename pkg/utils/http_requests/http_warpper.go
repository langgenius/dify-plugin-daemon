package http_requests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	routinepkg "github.com/langgenius/dify-plugin-daemon/pkg/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/stream"
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
	var watchdog *streamWatchdog
	var limits *StreamTimeouts
	ctx := context.Background()
	for _, option := range options {
		if option.Type == HttpOptionTypeContext {
			if value, ok := option.Value.(context.Context); ok {
				ctx = value
			}
		} else if option.Type == HttpOptionTypeStreamTimeouts {
			value := option.Value.(StreamTimeouts)
			limits = &value
		}
	}
	if limits != nil {
		var err error
		watchdog, err = newStreamWatchdog(ctx, *limits)
		if err != nil {
			return nil, err
		}
		options = append(options, HttpContext(watchdog.ctx))
	}
	resp, err := Request(client, url, method, options...)
	if err != nil {
		if watchdog != nil {
			if cause := watchdog.stop(); cause != nil {
				err = cause
			}
		}
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		if watchdog != nil {
			defer watchdog.stop()
		}
		errorText, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status code: %d and respond with: %s", resp.StatusCode, errorText)
	}

	ch := stream.NewStream[T](1024)

	// get read timeout
	readTimeout := int64(60000)
	raiseErrorWhenStreamDataNotMatch := false
	usingLengthPrefixed := false
	maxChunkSize := int64(1024 * 1024 * 30)
	for _, option := range options {
		if option.Type == HttpOptionTypeReadTimeout {
			readTimeout = option.Value.(int64)
		} else if option.Type == HttpOptionTypeRaiseErrorWhenStreamDataNotMatch {
			raiseErrorWhenStreamDataNotMatch = option.Value.(bool)
		} else if option.Type == HttpOptionTypeUsingLengthPrefixed {
			usingLengthPrefixed = option.Value.(bool)
		} else if option.Type == HttpOptionTypeMaxChunkSize {
			maxChunkSize = option.Value.(int64)
		}
	}
	var legacyTimer *time.Timer
	if watchdog == nil {
		legacyTimer = time.AfterFunc(time.Millisecond*time.Duration(readTimeout), func() {
			resp.Body.Close()
		})
	}
	ch.OnClose(func() {
		if watchdog != nil {
			watchdog.stop()
		}
		if legacyTimer != nil {
			legacyTimer.Stop()
		}
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
				log.Warn("stream data not match", "url", url, "data", string(data))
				return nil
			}
		}

		if watchdog != nil {
			if err := watchdog.progress(true); err != nil {
				return err
			}
		}
		ch.Write(t)
		return nil
	}

	routine.Submit(routinepkg.Labels{
		routinepkg.RoutineLabelKeyModule: "http_requests",
		routinepkg.RoutineLabelKeyMethod: "RequestAndParseStream",
	}, func() {
		defer ch.Close()
		var reader io.Reader = resp.Body
		if watchdog != nil {
			reader = &streamActivityReader{Reader: reader, watchdog: watchdog}
		}

		var err error
		if usingLengthPrefixed {
			err = parser.LengthPrefixedChunking(reader, 0x0f, uint32(maxChunkSize), processData)
		} else {
			err = parser.LineBasedChunking(reader, int(maxChunkSize), func(data []byte) error {
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

		if watchdog != nil {
			if cause := watchdog.stop(); cause != nil {
				err = cause
			}
		}
		if err != nil {
			ch.WriteError(err)
		}
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
