package plugin_manager

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/serverless"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models/curd"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

// InstallToAWSFromPkg installs a plugin to AWS Lambda
func (p *PluginManager) InstallToAWSFromPkg(
	tenant_id string,
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
	unique_identity, err := decoder.UniqueIdentity()
	if err != nil {
		return nil, err
	}

	response, err := serverless.UploadPlugin(decoder)
	if err != nil {
		return nil, err
	}

	new_response := stream.NewStream[PluginInstallResponse](128)
	routine.Submit(func() {
		defer func() {
			new_response.Close()
		}()

		lambda_url := ""
		lambda_function_name := ""

		response.Async(func(r serverless.LaunchAWSLambdaFunctionResponse) {
			if r.Event == serverless.Info {
				new_response.Write(PluginInstallResponse{
					Event: PluginInstallEventInfo,
					Data:  "Installing...",
				})
			} else if r.Event == serverless.Done {
				if lambda_url == "" || lambda_function_name == "" {
					new_response.Write(PluginInstallResponse{
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
					serverless_model := &models.ServerlessRuntime{
						Checksum:               checksum,
						Type:                   models.SERVERLESS_RUNTIME_TYPE_AWS_LAMBDA,
						FunctionURL:            lambda_url,
						FunctionName:           lambda_function_name,
						PluginUniqueIdentifier: unique_identity.String(),
						Declaration:            declaration,
					}
					err = db.Create(serverless_model)
					if err != nil {
						new_response.Write(PluginInstallResponse{
							Event: PluginInstallEventError,
							Data:  "Failed to create serverless runtime",
						})
						return
					}
				} else if err != nil {
					new_response.Write(PluginInstallResponse{
						Event: PluginInstallEventError,
						Data:  "Failed to check if the plugin is already installed",
					})
					return
				}

				_, _, err = curd.InstallPlugin(
					tenant_id,
					unique_identity,
					plugin_entities.PLUGIN_RUNTIME_TYPE_AWS,
					&declaration,
					source,
					meta,
				)
				if err != nil {
					new_response.Write(PluginInstallResponse{
						Event: PluginInstallEventError,
						Data:  "Failed to create plugin",
					})
					return
				}

				new_response.Write(PluginInstallResponse{
					Event: PluginInstallEventDone,
					Data:  "Installed",
				})
			} else if r.Event == serverless.Error {
				new_response.Write(PluginInstallResponse{
					Event: PluginInstallEventError,
					Data:  "Internal server error",
				})
			} else if r.Event == serverless.LambdaUrl {
				lambda_url = r.Message
			} else if r.Event == serverless.Lambda {
				lambda_function_name = r.Message
			} else {
				new_response.WriteError(fmt.Errorf("unknown event: %s, with message: %s", r.Event, r.Message))
			}
		})
	})

	return new_response, nil
}
