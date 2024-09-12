package serverless

import (
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

var (
	SERVERLESS_CONNECTOR_API_KEY string
	baseurl                      *url.URL
	client                       *http.Client
)

func Init(config *app.Config) {
	var err error
	baseurl, err = url.Parse(*config.DifyPluginServerlessConnectorURL)
	if err != nil {
		log.Panic("Failed to parse serverless connector url", err)
	}

	client = &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   5 * time.Second,   // how long a http connection can be alive before it's closed
				KeepAlive: 120 * time.Second, // how long a real tcp connection can be idle before it's closed
			}).Dial,
			IdleConnTimeout: 120 * time.Second,
		},
	}

	SERVERLESS_CONNECTOR_API_KEY = *config.DifyPluginServerlessConnectorAPIKey

	if err := Ping(); err != nil {
		log.Panic("Failed to ping serverless connector", err)
	}

	log.Info("Serverless connector initialized")
}
