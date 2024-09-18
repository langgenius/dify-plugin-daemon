package real

import (
	"net/http"
	"net/url"
)

type RealBackwardsInvocation struct {
	PLUGIN_INNER_API_KEY string
	baseurl              *url.URL
	client               *http.Client
}
