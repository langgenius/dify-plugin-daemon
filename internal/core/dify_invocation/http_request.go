package dify_invocation

import "github.com/langgenius/dify-plugin-daemon/internal/utils/requests"

func Request[T any](method string, path string, options ...requests.HttpOptions) (*T, error) {
	return requests.RequestAndParse[T](client, difyPath(path), method, options...)
}
