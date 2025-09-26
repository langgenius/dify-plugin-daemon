package plugin_manager

import (
	"fmt"

	serverless "github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/serverless_connector"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

// InstallToServerlessFromPkg installs a plugin to Serverless
func (p *PluginManager) InstallToServerlessFromPkg(
	originalPackager []byte,
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
	// check valid manifest
	_, err = decoder.Manifest()
	if err != nil {
		return nil, err
	}
	uniqueIdentity, err := decoder.UniqueIdentity()
	if err != nil {
		return nil, err
	}

	// serverless.LaunchPlugin will check if the plugin has already been launched, if so, it returns directly
	response, err := serverless.LaunchPlugin(uniqueIdentity, originalPackager, decoder, p.config.DifyPluginServerlessConnectorLaunchTimeout, false)
	if err != nil {
		return nil, err
	}

	newResponse := stream.NewStream[PluginInstallResponse](128)
	routine.Submit(map[string]string{
		"module":          "plugin_manager",
		"function":        "InstallToServerlessFromPkg",
		"checksum":        checksum,
		"unique_identity": uniqueIdentity.String(),
		"source":          source,
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
	})

	return newResponse, nil
}

/*
 * Reinstall a plugin to Serverless, update function url and name
 */
func (p *PluginManager) ReinstallToServerlessFromPkg(
	originalPackager []byte,
	decoder decoder.PluginDecoder,
) (
	*stream.Stream[PluginInstallResponse], error,
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
		p.config.DifyPluginServerlessConnectorLaunchTimeout,
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

				// clear cache
				err = p.clearServerlessRuntimeCache(uniqueIdentity)
				if err != nil {
					newResponse.Write(PluginInstallResponse{
						Event: PluginInstallEventError,
						Data:  "Failed to clear serverless runtime cache",
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
