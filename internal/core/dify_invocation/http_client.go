package dify_invocation

import (
	"net"
	"net/http"
	"net/url"
	"time"
)

var (
	PLUGIN_INNER_API_KEY string
	baseurl              *url.URL
	client               *http.Client
)

func InitDifyInvocationDaemon(base string, calling_key string) error {
	var err error
	baseurl, err = url.Parse(base)
	if err != nil {
		return err
	}

	client = &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 15 * time.Second,
			}).Dial,
			IdleConnTimeout: 120 * time.Second,
		},
	}

	PLUGIN_INNER_API_KEY = calling_key

	return nil
}
