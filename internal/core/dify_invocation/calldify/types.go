package calldify

import (
	"context"
	"net/http"
	"net/url"
)

type RealBackwardsInvocation struct {
	difyInnerApiKey     string
	difyInnerApiBaseurl *url.URL
	client              *http.Client
	writeTimeout        int64
	readTimeout         int64
	traceCtx            context.Context
}

func (i *RealBackwardsInvocation) SetContext(ctx context.Context) {
	i.traceCtx = ctx
}

func (i *RealBackwardsInvocation) Context() context.Context {
	if i.traceCtx == nil {
		return context.Background()
	}
	return i.traceCtx
}

type BaseBackwardsInvocationResponse[T any] struct {
	Data  *T     `json:"data,omitempty"`
	Error string `json:"error"`
}
