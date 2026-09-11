package calldify

import (
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/http_requests"
	otelhttp "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type NewDifyInvocationDaemonPayload struct {
	BaseUrl                 string
	CallingKey              string
	WriteTimeout            int64
	ReadTimeout             int64
	LLMFirstResponseTimeout int64
	LLMIdleTimeout          int64
	LLMTotalTimeout         int64
	ResponseMaxBufferSize   int64
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
	total := payload.LLMTotalTimeout
	if total == 0 {
		total = payload.ReadTimeout
	}
	firstResponse := payload.LLMFirstResponseTimeout
	if firstResponse == 0 {
		firstResponse = total
	}
	idle := payload.LLMIdleTimeout
	if idle == 0 {
		idle = payload.ReadTimeout
	}
	invocation.llmStreamTimeouts = http_requests.StreamTimeouts{
		FirstResponse: time.Duration(firstResponse) * time.Millisecond,
		ReadIdle:      time.Duration(idle) * time.Millisecond,
		Total:         time.Duration(total) * time.Millisecond,
	}
	invocation.responseMaxBufferSize = payload.ResponseMaxBufferSize

	return invocation, nil
}
