package serverless_runtime

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/basic_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/media_transport"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/mapping"
)

type ServerlessPluginRuntime struct {
	basic_runtime.BasicChecksum
	plugin_entities.PluginRuntime

	// access url for the lambda function
	LambdaURL  string
	LambdaName string

	// listeners mapping session id to the listener
	listeners mapping.Map[string, *entities.Broadcast[plugin_entities.SessionMessage]]

	Client *http.Client

	PluginMaxExecutionTimeout int // in seconds

	RuntimeBufferSize    int
	RuntimeMaxBufferSize int
}

// build a serverless plugin runtime
func ConstructServerlessPluginRuntime(
	config *app.Config,
	pluginDeclaration *plugin_entities.PluginDeclaration,
	serverlessModel *models.ServerlessRuntime,
	mediaBucket *media_transport.MediaBucket,
	plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
) *ServerlessPluginRuntime {
	// init runtime entity
	runtimeEntity := plugin_entities.PluginRuntime{
		Config: *pluginDeclaration,
	}
	runtimeEntity.InitState()

	return &ServerlessPluginRuntime{
		BasicChecksum: basic_runtime.BasicChecksum{
			MediaTransport: basic_runtime.NewMediaTransport(mediaBucket),
			InnerChecksum:  serverlessModel.Checksum,
		},
		PluginRuntime:             runtimeEntity,
		LambdaURL:                 serverlessModel.FunctionURL,
		LambdaName:                serverlessModel.FunctionName,
		PluginMaxExecutionTimeout: config.PluginMaxExecutionTimeout,
		RuntimeBufferSize:         config.PluginRuntimeBufferSize,
		RuntimeMaxBufferSize:      config.PluginRuntimeMaxBufferSize,

		// init http client
		Client: &http.Client{
			Transport: &http.Transport{
				TLSHandshakeTimeout: time.Duration(config.PluginMaxExecutionTimeout) * time.Second,
				IdleConnTimeout:     120 * time.Second,
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					conn, err := (&net.Dialer{
						Timeout:   time.Duration(config.PluginMaxExecutionTimeout) * time.Second,
						KeepAlive: 120 * time.Second,
					}).DialContext(ctx, network, addr)
					if err != nil {
						return nil, err
					}
					return conn, nil
				},
			},
		},
	}
}
