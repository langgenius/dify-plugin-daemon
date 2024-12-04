package plugin_manager

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/serverless"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

// InstallToAWSFromPkg installs a plugin to AWS Lambda
func (p *PluginManager) InstallToAWSFromPkg(
	decoder decoder.PluginDecoder,
	source string,
	meta map[string]any,
) (
	*stream.Stream[PluginInstallResponse], error,
) {
	checksum, err := decoder.Checksum()
	if err != nil {
		return nil, err
	}
	declaration, err := decoder.Manifest()
	if err != nil {
		return nil, err
	}
	uniqueIdentity, err := decoder.UniqueIdentity()
	if err != nil {
		return nil, err
	}

	response, err := serverless.UploadPlugin(decoder)
	if err != nil {
		return nil, err
	}

	newResponse := stream.NewStream[PluginInstallResponse](128)
	routine.Submit(map[string]string{
		"module":          "plugin_manager",
		"function":        "InstallToAWSFromPkg",
		"checksum":        checksum,
		"unique_identity": uniqueIdentity.String(),
		"source":          source,
	}, func() {
		defer func() {
			newResponse.Close()
		}()

		lambdaUrl := ""
		lambdaFunctionName := ""

		response.Async(func(r serverless.LaunchAWSLambdaFunctionResponse) {
			if r.Event == serverless.Info {
				newResponse.Write(PluginInstallResponse{
					Event: PluginInstallEventInfo,
					Data:  "Installing...",
				})
			} else if r.Event == serverless.Done {
				if lambdaUrl == "" || lambdaFunctionName == "" {
					newResponse.Write(PluginInstallResponse{
						Event: PluginInstallEventError,
						Data:  "Internal server error, failed to get lambda url or function name",
					})
					return
				}
				// check if the plugin is already installed
				_, err := db.GetOne[models.ServerlessRuntime](
					db.Equal("checksum", checksum),
					db.Equal("type", string(models.SERVERLESS_RUNTIME_TYPE_AWS_LAMBDA)),
				)
				if err == db.ErrDatabaseNotFound {
					// create a new serverless runtime
					serverlessModel := &models.ServerlessRuntime{
						Checksum:               checksum,
						Type:                   models.SERVERLESS_RUNTIME_TYPE_AWS_LAMBDA,
						FunctionURL:            lambdaUrl,
						FunctionName:           lambdaFunctionName,
						PluginUniqueIdentifier: uniqueIdentity.String(),
						Declaration:            declaration,
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
			} else if r.Event == serverless.LambdaUrl {
				lambdaUrl = r.Message
			} else if r.Event == serverless.Lambda {
				lambdaFunctionName = r.Message
			} else {
				newResponse.WriteError(fmt.Errorf("unknown event: %s, with message: %s", r.Event, r.Message))
			}
		})
	})

	return newResponse, nil
}
