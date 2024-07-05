package dify_invocation

import (
	"net"
	"net/http"
	"net/url"
	"time"
)

var (
	baseurl *url.URL
	client  *http.Client
)

func InitDifyInvocationDaemon(base string) error {
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

	return nil
}
