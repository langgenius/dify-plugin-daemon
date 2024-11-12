package real

import (
	"net/http"
	"net/url"
)

type RealBackwardsInvocation struct {
	difyInnerApiKey     string
	difyInnerApiBaseurl *url.URL
	client              *http.Client
}

type BaseBackwardsInvocationResponse[T any] struct {
	Data  *T     `json:"data,omitempty"`
	Error string `json:"error"`
}
