package serverless

import (
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
	otelhttp "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var (
	SERVERLESS_CONNECTOR_API_KEY string
	baseurl                      *url.URL
	client                       *http.Client
)

func Init(config *app.Config) {
	ctx := log.EnsureTrace(nil)

	var err error
	baseurl, err = url.Parse(*config.DifyPluginServerlessConnectorURL)
	if err != nil {
		log.PanicContext(ctx, "Failed to parse serverless connector url", "error", err)
	}

	client = &http.Client{
		Transport: otelhttp.NewTransport(&http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,   // how long a http connection can be alive before it's closed
				KeepAlive: 120 * time.Second, // how long a real tcp connection can be idle before it's closed
			}).DialContext,
			IdleConnTimeout: 120 * time.Second,
		}),
	}
	SERVERLESS_CONNECTOR_API_KEY = *config.DifyPluginServerlessConnectorAPIKey

	if err := PingWithContext(ctx); err != nil {
		log.PanicContext(ctx, "Failed to ping serverless connector", "error", err)
	}

	log.InfoContext(ctx, "Serverless connector initialized")
}
