package controlpanel

import (
	"fmt"

	serverless "github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/serverless_connector"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

func (c *ControlPanel) InstallToServerlessFromPkg(
	packageFile []byte,
	decoder decoder.PluginDecoder,
) (
	*stream.Stream[InstallServerlessPluginResponse], error,
) {
	// check valid manifest
	_, err := decoder.Manifest()
	if err != nil {
		return nil, err
	}

	uniqueIdentity, err := decoder.UniqueIdentity()
	if err != nil {
		return nil, err
	}

	// serverless.LaunchPlugin will check if the plugin has already been launched, if so, it returns directly
	response, err := serverless.LaunchPlugin(
		uniqueIdentity,
		packageFile,
		decoder,
		c.config.DifyPluginServerlessConnectorLaunchTimeout,
		false,
	)
	if err != nil {
		return nil, err
	}

	return response, nil
}

/*
 * Reinstall a plugin to Serverless, update function url and name
 */
func (c *ControlPanel) ReinstallToServerlessFromPkg(
	originalPackager []byte,
	decoder decoder.PluginDecoder,
) (
	*stream.Stream[InstallServerlessPluginResponse], error,
) {
	checksum, err := decoder.Checksum()
	if err != nil {
		return nil, err
	}
	// check valid manifest
	_, err = decoder.Manifest()
	if err != nil {
		return nil, err
	}
	uniqueIdentity, err := decoder.UniqueIdentity()
	if err != nil {
		return nil, err
	}

	// check if serverless runtime exists
	serverlessRuntime, err := db.GetOne[models.ServerlessRuntime](
		db.Equal("plugin_unique_identifier", uniqueIdentity.String()),
	)
	if err == db.ErrDatabaseNotFound {
		return nil, fmt.Errorf("plugin not exists")
	}
	if err != nil {
		return nil, err
	}

	response, err := serverless.LaunchPlugin(
		uniqueIdentity,
		originalPackager,
		decoder,
		c.config.DifyPluginServerlessConnectorLaunchTimeout,
		true, // ignoreIdempotent, true means always reinstall
	)
	if err != nil {
		return nil, err
	}

	newResponse := stream.NewStream[PluginInstallResponse](128)
	routine.Submit(map[string]string{
		"module":          "plugin_manager",
		"function":        "ReinstallToServerlessFromPkg",
		"checksum":        checksum,
		"unique_identity": uniqueIdentity.String(),
	}, func() {
		defer func() {
			newResponse.Close()
		}()

		functionUrl := ""
		functionName := ""

		response.Async(func(r serverless.LaunchFunctionResponse) {
			if r.Event == serverless.Info {
				newResponse.Write(PluginInstallResponse{
					Event: PluginInstallEventInfo,
					Data:  "Installing...",
				})
			} else if r.Event == serverless.Done {
				if functionUrl == "" || functionName == "" {
					newResponse.Write(PluginInstallResponse{
						Event: PluginInstallEventError,
						Data:  "Internal server error, failed to get lambda url or function name",
					})
					return
				}

				// update serverless runtime
				serverlessRuntime.FunctionURL = functionUrl
				serverlessRuntime.FunctionName = functionName
				err = db.Update(&serverlessRuntime)
				if err != nil {
					newResponse.Write(PluginInstallResponse{
						Event: PluginInstallEventError,
						Data:  "Failed to update serverless runtime",
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
	})

	return newResponse, nil
}
