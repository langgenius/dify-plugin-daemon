package real

import (
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
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
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 120 * time.Second,
			}).Dial,
			IdleConnTimeout: 120 * time.Second,
		},
	}

	invocation.difyInnerApiBaseurl = baseurl
	invocation.client = client
	invocation.difyInnerApiKey = payload.CallingKey
	invocation.writeTimeout = payload.WriteTimeout
	invocation.readTimeout = payload.ReadTimeout

	return invocation, nil
}
