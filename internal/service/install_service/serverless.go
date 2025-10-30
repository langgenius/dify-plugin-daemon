package install_service

import (
	"fmt"

	serverless "github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/serverless_connector"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

func HandleServerlessInstallationStream(
	pkgFile []byte,
	pluginDecoder decoder.PluginDecoder,
	stream *stream.Stream[serverless.LaunchFunctionResponse],
) (*stream.Stream[PluginInstallResponse], error) {
	functionUrl := ""
	functionName := ""

	response.Async(func(r serverless.LaunchFunctionResponse) {
		if r.Event == serverless.Info {
			newResponse.Write(InstallServerlessPluginResponse{
				Event:   serverless.Info,
				Message: "Installing...",
			})
		} else if r.Event == serverless.Done {
			if functionUrl == "" || functionName == "" {
				newResponse.Write(InstallServerlessPluginResponse{
					Event:   serverless.Error,
					Message: "Internal server error, failed to get lambda url or function name",
				})
				return
			}
			// check if the plugin is already installed
			_, err := db.GetOne[models.ServerlessRuntime](
				db.Equal("checksum", checksum),
				db.Equal("type", string(models.SERVERLESS_RUNTIME_TYPE_SERVERLESS)),
			)
			if err == db.ErrDatabaseNotFound {
				// create a new serverless runtime
				serverlessModel := &models.ServerlessRuntime{
					Checksum:               checksum,
					Type:                   models.SERVERLESS_RUNTIME_TYPE_SERVERLESS,
					FunctionURL:            functionUrl,
					FunctionName:           functionName,
					PluginUniqueIdentifier: uniqueIdentity.String(),
				}
				err = db.Create(serverlessModel)
				if err != nil {
					newResponse.Write(PluginInstallResponse{
						Event: PluginInstallEventError,
						Data:  "Failed to create serverless runtime",
					})
					return
				}
			} else if err != nil {
				newResponse.Write(PluginInstallResponse{
					Event: PluginInstallEventError,
					Data:  "Failed to check if the plugin is already installed",
				})
				return
			}

			newResponse.Write(PluginInstallResponse{
				Event: PluginInstallEventDone,
				Data:  "Installed",
			})
		} else if r.Event == serverless.Error {
			newResponse.Write(PluginInstallResponse{
				Event: PluginInstallEventError,
				Data:  "Internal server error",
			})
		} else if r.Event == serverless.FunctionUrl {
			functionUrl = r.Message
		} else if r.Event == serverless.Function {
			functionName = r.Message
		} else {
			newResponse.WriteError(fmt.Errorf("unknown event: %s, with message: %s", r.Event, r.Message))
		}
	})
	return nil, nil
}
