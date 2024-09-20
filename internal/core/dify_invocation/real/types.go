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
