package calldify

import (
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	otelhttp "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type NewDifyInvocationDaemonPayload struct {
	BaseUrl      string
	CallingKey   string
	WriteTimeout int64
	ReadTimeout  int64
}

func NewDifyInvocationDaemon(payload NewDifyInvocationDaemonPayload) (dify_invocation.BackwardsInvocation, error) {
	var err error
	invocation := &RealBackwardsInvocation{}
	baseurl, err := url.Parse(payload.BaseUrl)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: otelhttp.NewTransport(&http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 120 * time.Second,
			}).DialContext,
			IdleConnTimeout: 120 * time.Second,
		}),
	}
	invocation.difyInnerApiBaseurl = baseurl
	invocation.client = client
	invocation.difyInnerApiKey = payload.CallingKey
	invocation.writeTimeout = payload.WriteTimeout
	invocation.readTimeout = payload.ReadTimeout

	return invocation, nil
}
