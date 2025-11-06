package plugin_manager

import (
	"errors"
	"fmt"

	controlpanel "github.com/langgenius/dify-plugin-daemon/internal/core/control_panel"
	serverless "github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/serverless_connector"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/installation_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

var (
	ErrReinstallNotSupported = errors.New("reinstall is not supported on local platform")
)

func (p *PluginManager) Install(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (*stream.Stream[installation_entities.PluginInstallResponse], error) {
	if p.config.Platform == app.PLATFORM_LOCAL {
		return p.installLocal(pluginUniqueIdentifier)
	}

	return p.installServerless(pluginUniqueIdentifier)
}

func (p *PluginManager) Reinstall(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (*stream.Stream[installation_entities.PluginInstallResponse], error) {
	if p.config.Platform == app.PLATFORM_LOCAL {
		return nil, ErrReinstallNotSupported
	}

	// TODO: implement reinstall for serverless platform
	return nil, nil
}

func (p *PluginManager) installServerless(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (*stream.Stream[installation_entities.PluginInstallResponse], error) {
	response, err := p.controlPanel.InstallToServerlessFromPkg(pluginUniqueIdentifier)
	if err != nil {
		return nil, errors.Join(
			errors.New("failed to install plugin to serverless"),
			err,
		)
	}

	functionUrl := ""
	functionName := ""

	responseStream := stream.NewStream[installation_entities.PluginInstallResponse](128)
	response.Async(func(r serverless.LaunchFunctionResponse) {
		if r.Event == serverless.Info {
			responseStream.Write(installation_entities.PluginInstallResponse{
				Event: installation_entities.PluginInstallEventInfo,
				Data:  "Installing...",
			})
		} else if r.Event == serverless.Done {
			if functionUrl == "" || functionName == "" {
				responseStream.Write(installation_entities.PluginInstallResponse{
					Event: installation_entities.PluginInstallEventError,
					Data:  "Internal server error, failed to get lambda url or function name",
				})
				return
			}
			// check if the plugin is already installed
			_, err := db.GetOne[models.ServerlessRuntime](
				db.Equal("plugin_unique_identifier", pluginUniqueIdentifier.String()),
				db.Equal("type", string(models.SERVERLESS_RUNTIME_TYPE_SERVERLESS)),
			)
			if err == db.ErrDatabaseNotFound {
				// create a new serverless runtime
				serverlessModel := &models.ServerlessRuntime{
					Checksum:               pluginUniqueIdentifier.Checksum(),
					Type:                   models.SERVERLESS_RUNTIME_TYPE_SERVERLESS,
					FunctionURL:            functionUrl,
					FunctionName:           functionName,
					PluginUniqueIdentifier: pluginUniqueIdentifier.String(),
				}
				err = db.Create(serverlessModel)
				if err != nil {
					responseStream.Write(installation_entities.PluginInstallResponse{
						Event: installation_entities.PluginInstallEventError,
						Data:  "Failed to create serverless runtime",
					})
					return
				}
			} else if err != nil {
				responseStream.Write(installation_entities.PluginInstallResponse{
					Event: installation_entities.PluginInstallEventError,
					Data:  "Failed to check if the plugin is already installed",
				})
				return
			}

			responseStream.Write(installation_entities.PluginInstallResponse{
				Event: installation_entities.PluginInstallEventDone,
				Data:  "Installed",
			})
		} else if r.Event == serverless.Error {
			responseStream.Write(installation_entities.PluginInstallResponse{
				Event: installation_entities.PluginInstallEventError,
				Data:  "Internal server error",
			})
		} else if r.Event == serverless.FunctionUrl {
			functionUrl = r.Message
		} else if r.Event == serverless.Function {
			functionName = r.Message
		} else {
			responseStream.WriteError(fmt.Errorf("unknown event: %s, with message: %s", r.Event, r.Message))
		}
	})

	return responseStream, nil
}

func (p *PluginManager) installLocal(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (*stream.Stream[installation_entities.PluginInstallResponse], error) {
	responseStream := stream.NewStream[installation_entities.PluginInstallResponse](128)

	response, err := p.controlPanel.InstallToLocalFromPkg(pluginUniqueIdentifier)
	if err != nil {
		return nil, errors.Join(
			errors.New("failed to install plugin to local"),
			err,
		)
	}

	routine.Submit(map[string]string{
		"module": "plugin_manager",
		"action": "installLocal",
	}, func() {
		defer responseStream.Close()
		response.Async(func(response controlpanel.InstallLocalPluginResponse) {
			if response.Event == installation_entities.PluginInstallEventInfo {
				responseStream.Write(installation_entities.PluginInstallResponse{
					Event: installation_entities.PluginInstallEventInfo,
					Data:  "Installing...",
				})
			}
			if response.Event == installation_entities.PluginInstallEventDone {
				responseStream.Write(installation_entities.PluginInstallResponse{
					Event: installation_entities.PluginInstallEventDone,
					Data:  "Installed",
				})
			}
			if response.Event == installation_entities.PluginInstallEventError {
				responseStream.Write(installation_entities.PluginInstallResponse{
					Event: installation_entities.PluginInstallEventError,
					Data:  "Failed to install plugin to local",
				})
			}
		})
	})

	return responseStream, nil
}
