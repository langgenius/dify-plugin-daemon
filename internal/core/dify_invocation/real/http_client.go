package real

import (
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
)

func NewDifyInvocationDaemon(base string, calling_key string) (dify_invocation.BackwardsInvocation, error) {
	var err error
	invocation := &RealBackwardsInvocation{}
	baseurl, err := url.Parse(base)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 120 * time.Second,
			}).Dial,
			IdleConnTimeout: 120 * time.Second,
		},
	}

	invocation.dify_inner_api_baseurl = baseurl
	invocation.client = client
	invocation.dify_inner_api_key = calling_key

	return invocation, nil
}
