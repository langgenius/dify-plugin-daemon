package aws_manager

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

func (r *AWSPluginRuntime) InitEnvironment() error {
	// init http client
	r.client = &http.Client{
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

func (r *AWSPluginRuntime) Identity() (plugin_entities.PluginUniqueIdentifier, error) {
	return plugin_entities.PluginUniqueIdentifier(fmt.Sprintf("%s@%s", r.Config.Identity(), r.Checksum())), nil
}
