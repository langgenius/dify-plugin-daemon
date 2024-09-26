package real

import (
	"net/http"
	"net/url"
)

type RealBackwardsInvocation struct {
	dify_inner_api_key     string
	dify_inner_api_baseurl *url.URL
	client                 *http.Client
}

type BaseBackwardsInvocationResponse[T any] struct {
	Data  *T     `json:"data,omitempty"`
	Error string `json:"error"`
}
