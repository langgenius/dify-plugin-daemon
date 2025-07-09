package serverless_runtime

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func (r *ServerlessPluginRuntime) InitEnvironment() error {
	// init http client
	r.client = &http.Client{
		Transport: &http.Transport{
			TLSHandshakeTimeout: time.Duration(r.PluginMaxExecutionTimeout) * time.Second,
			IdleConnTimeout:     120 * time.Second,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				conn, err := (&net.Dialer{
					Timeout:   time.Duration(r.PluginMaxExecutionTimeout) * time.Second,
					KeepAlive: 120 * time.Second,
				}).DialContext(ctx, network, addr)
				if err != nil {
					return nil, err
				}
				return conn, nil
			},
		},
	}

	return nil
}

func (r *ServerlessPluginRuntime) Identity() (plugin_entities.PluginUniqueIdentifier, error) {
	checksum, err := r.Checksum()
	if err != nil {
		return "", err
	}
	return plugin_entities.NewPluginUniqueIdentifier(fmt.Sprintf("%s@%s", r.Config.Identity(), checksum))
}
