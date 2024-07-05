package requests

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

func parseJsonBody(resp *http.Response, ret interface{}) error {
	defer resp.Body.Close()
	json_decoder := json.NewDecoder(resp.Body)
	return json_decoder.Decode(ret)
}

func RequestAndParse[T any](client *http.Client, url string, method string, options ...HttpOptions) (*T, error) {
	var ret T

	resp, err := Request(client, url, method, options...)
	if err != nil {
		return nil, err
	}

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

func RequestAndParseStream[T any](client *http.Client, url string, method string, options ...HttpOptions) (*stream.StreamResponse[T], error) {
	resp, err := Request(client, url, method, options...)
	if err != nil {
		return nil, err
	}

	ch := stream.NewStreamResponse[T](1024)

	routine.Submit(func() {
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			data := scanner.Bytes()
			if bytes.HasPrefix(data, []byte("data: ")) {
				// split
				data = data[6:]
				// unmarshal
				t, err := parser.UnmarshalJsonBytes[T](data)
				if err != nil {
					continue
				}

				ch.Write(t)
			}
		}

		ch.Close()
	})

	return ch, nil
}

func GetAndParseStream[T any](client *http.Client, url string, options ...HttpOptions) (*stream.StreamResponse[T], error) {
	return RequestAndParseStream[T](client, url, "GET", options...)
}

func PostAndParseStream[T any](client *http.Client, url string, options ...HttpOptions) (*stream.StreamResponse[T], error) {
	return RequestAndParseStream[T](client, url, "POST", options...)
}

func PutAndParseStream[T any](client *http.Client, url string, options ...HttpOptions) (*stream.StreamResponse[T], error) {
	return RequestAndParseStream[T](client, url, "PUT", options...)
}
