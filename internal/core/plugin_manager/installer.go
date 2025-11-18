package plugin_manager

import (
	"errors"
	"fmt"
	"time"

	controlpanel "github.com/langgenius/dify-plugin-daemon/internal/core/control_panel"
	"github.com/langgenius/dify-plugin-daemon/internal/core/local_runtime"
	serverless "github.com/langgenius/dify-plugin-daemon/internal/core/serverless_connector"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/installation_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	routinepkg "github.com/langgenius/dify-plugin-daemon/pkg/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/stream"
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

// serverless runtime uses a strategy that firstly compile the plugin into a docker image
// then execute it on a docker container(k8s pod / aws lambda)
// however, it's mutable due to dependencies updates when using a version range to
// limit the dependencies, e.g. `dify_plugin>=1.0.0,<2.0.0`
//
// but serverless runtime persists the image, that's why we introduced `Reinstall`
// it recompiles the plugin and launch a new plugin runtime, replace the old one
func (p *PluginManager) Reinstall(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (*stream.Stream[installation_entities.PluginInstallResponse], error) {
	if p.config.Platform == app.PLATFORM_LOCAL {
		return nil, ErrReinstallNotSupported
	}

	response, err := p.controlPanel.ReinstallToServerless(pluginUniqueIdentifier)
	if err != nil {
		return nil, errors.Join(
			errors.New("failed to reinstall plugin to serverless"),
			err,
		)
	}

	responseStream := stream.NewStream[installation_entities.PluginInstallResponse](128)

	routine.Submit(routinepkg.Labels{
		routinepkg.RoutineLabelKeyModule: "plugin_manager",
		routinepkg.RoutineLabelKeyMethod: "reinstallServerless",
	}, func() {
		defer responseStream.Close()

		functionUrl := ""
		functionName := ""

		if err := response.Process(func(ispr serverless.LaunchFunctionResponse) {
			switch ispr.Event {
			case serverless.Done:
				if functionUrl == "" || functionName == "" {
					responseStream.Write(installation_entities.PluginInstallResponse{
						Event: installation_entities.PluginInstallEventError,
						Data:  "Internal server error, failed to get serverless function url or function name",
					})
					return
				}

				// update serverless runtime model
				if err := p.updateServerlessRuntimeModel(pluginUniqueIdentifier, functionUrl, functionName); err != nil {
					responseStream.Write(installation_entities.PluginInstallResponse{
						Event: installation_entities.PluginInstallEventError,
						Data:  "failed to get serverless runtime model",
					})
					return
				}

				// cleanup system cache for serverless runtime model
				// cleanup must be done after updating the model, otherwise race condition may occur
				if err := p.clearServerlessRuntimeCache(pluginUniqueIdentifier); err != nil {
					log.Error("failed to cleanup system cache for serverless runtime model: %v", err)
					responseStream.Write(installation_entities.PluginInstallResponse{
						Event: installation_entities.PluginInstallEventError,
						Data:  "failed to cleanup system cache for serverless runtime model",
					})
					return
				}

				responseStream.Write(installation_entities.PluginInstallResponse{
					Event: installation_entities.PluginInstallEventDone,
					Data:  "successfully reinstalled",
				})
			case serverless.Error:
				// FIXME(Yeuoly): log the error to terminal, but avoid using inline log
				// try to refactor the code to a more elegant way like abstracting all lifetime events
				// and make logger in a centralized layer
				log.Error("failed to reinstall plugin to serverless: %s", ispr.Message)
				responseStream.Write(installation_entities.PluginInstallResponse{
					Event: installation_entities.PluginInstallEventError,
					Data:  "failed to reinstall plugin to serverless",
				})
			case serverless.Info:
				responseStream.Write(installation_entities.PluginInstallResponse{
					Event: installation_entities.PluginInstallEventInfo,
					Data:  "reinstalling...",
				})
			case serverless.Function:
				functionName = ispr.Message
			case serverless.FunctionUrl:
				functionUrl = ispr.Message
			default:
				responseStream.WriteError(fmt.Errorf("unknown event: %s, with message: %s", ispr.Event, ispr.Message))
			}
		}); err != nil {
			responseStream.WriteError(err)
		}
	})

	return responseStream, nil
}

func (p *PluginManager) updateServerlessRuntimeModel(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	functionUrl string,
	functionName string,
) error {
	serverlessModel, err := db.GetOne[models.ServerlessRuntime](
		db.Equal("plugin_unique_identifier", pluginUniqueIdentifier.String()),
		db.Equal("type", string(models.SERVERLESS_RUNTIME_TYPE_SERVERLESS)),
	)
	if err != nil {
		return err
	}

	serverlessModel.FunctionURL = functionUrl
	serverlessModel.FunctionName = functionName

	return db.Update(&serverlessModel)
}

// whenever a plugin was installed successfully, a record will be inserted into `models.ServerlessRuntime`
func (p *PluginManager) installServerless(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (*stream.Stream[installation_entities.PluginInstallResponse], error) {
	response, err := p.controlPanel.InstallToServerless(pluginUniqueIdentifier)
	if err != nil {
		return nil, errors.Join(
			errors.New("failed to install plugin to serverless"),
			err,
		)
	}

	responseStream := stream.NewStream[installation_entities.PluginInstallResponse](128)

	routine.Submit(routinepkg.Labels{
		routinepkg.RoutineLabelKeyModule: "plugin_manager",
		routinepkg.RoutineLabelKeyMethod: "installServerless",
	}, func() {
		defer responseStream.Close()

		functionUrl := ""
		functionName := ""

		if err := response.Process(func(r serverless.LaunchFunctionResponse) {
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
				// NOTE: models.ServerlessRuntime is a tenant-isolated model
				// it hands only engine-level persist data like which serverless runtime is installed
				// that's why we placed it here, not in service layer.
				//
				// service layer takes care of tenant-level persist data like "which tenant installed which plugin"
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
							Data:  "failed to create serverless runtime",
						})
						return
					}
				} else if err != nil {
					responseStream.Write(installation_entities.PluginInstallResponse{
						Event: installation_entities.PluginInstallEventError,
						Data:  "failed to check if the plugin is already installed",
					})
					return
				}

				responseStream.Write(installation_entities.PluginInstallResponse{
					Event: installation_entities.PluginInstallEventDone,
					Data:  "successfully installed",
				})
			} else if r.Event == serverless.Error {
				// FIXME(Yeuoly): log the error to terminal, but avoid using inline log
				// try to refactor the code to a more elegant way like abstracting all lifetime events
				// and make logger in a centralized layer
				log.Error("failed to install plugin to serverless: %s", r.Message)
				responseStream.Write(installation_entities.PluginInstallResponse{
					Event: installation_entities.PluginInstallEventError,
					Data:  "internal server error",
				})
			} else if r.Event == serverless.FunctionUrl {
				functionUrl = r.Message
			} else if r.Event == serverless.Function {
				functionName = r.Message
			} else {
				responseStream.WriteError(fmt.Errorf("unknown event: %s, with message: %s", r.Event, r.Message))
			}
		}); err != nil {
			responseStream.WriteError(err)
		}
	})

	return responseStream, nil
}

func (p *PluginManager) installLocal(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (*stream.Stream[installation_entities.PluginInstallResponse], error) {
	responseStream := stream.NewStream[installation_entities.PluginInstallResponse](128)

	routine.Submit(routinepkg.Labels{
		routinepkg.RoutineLabelKeyModule: "plugin_manager",
		routinepkg.RoutineLabelKeyMethod: "installLocal",
	}, func() {
		// firstly, install the plugin, then launch it, delete it if process fails
		var success bool = false
		var runtime *local_runtime.LocalPluginRuntime
		var ch <-chan error

		defer responseStream.Close()
		defer func() {
			if !success {
				p.controlPanel.RemoveLocalPlugin(pluginUniqueIdentifier)

				// release the lock, avoid a potential race condition
				// which causes plugins never to be scheduled automatically
				p.controlPanel.EnableLocalPluginAutoLaunch(pluginUniqueIdentifier)

				// forcefully stop runtime, prevent continuous scheduling
				if runtime != nil {
					runtime.Stop(false)
				}
			}
		}()

		// to avoid race condition:
		// 	  WatchDog` starts the plugin before `installLocal` called `LaunchLocalPlugin`
		//    firstly disable the auto launch
		p.controlPanel.DisableLocalPluginAutoLaunch(pluginUniqueIdentifier)

		// move the plugin to installed bucket
		err := p.controlPanel.InstallToLocal(pluginUniqueIdentifier)
		if err != nil {
			responseStream.Write(installation_entities.PluginInstallResponse{
				Event: installation_entities.PluginInstallEventError,
				Data:  fmt.Sprintf("failed to move plugin to installed bucket: %s", err.Error()),
			})
			return
		}

		// call `LaunchLocalPlugin` to launch the plugin
		// `ch` is used to wait for the plugin to be ready or failed
		runtime, ch, err = p.controlPanel.LaunchLocalPlugin(pluginUniqueIdentifier)

		// if the plugin is already launched, just return success
		if err == controlpanel.ErrorPluginAlreadyLaunched {
			success = true // set success to true as the result for defer function
			responseStream.Write(installation_entities.PluginInstallResponse{
				Event: installation_entities.PluginInstallEventDone,
				Data:  "successfully installed",
			})
			return
		} else if err != nil {
			// release the lock
			responseStream.Write(installation_entities.PluginInstallResponse{
				Event: installation_entities.PluginInstallEventError,
				Data:  fmt.Sprintf("failed to launch plugin: %s", err.Error()),
			})
			return
		}

		ticker := time.NewTicker(5 * time.Second)
		timeout := time.Duration(p.config.PythonEnvInitTimeout) * time.Second
		timer := time.NewTimer(timeout)

		for {
			select {
			case <-timer.C:
				responseStream.Write(installation_entities.PluginInstallResponse{
					Event: installation_entities.PluginInstallEventError,
					Data: fmt.Sprintf(
						"timed out on waiting for plugin to be ready after %s, "+
							"please contract the administrator to check the logs",
						timeout.String(),
					),
				})
				return
			case <-ticker.C:
				// keep sending heartbeat until the plugin is ready or timed out
				responseStream.Write(installation_entities.PluginInstallResponse{
					Event: installation_entities.PluginInstallEventInfo,
					Data:  "installing heartbeat, waiting for plugin to be ready...",
				})
			case err := <-ch:
				if err != nil {
					responseStream.Write(installation_entities.PluginInstallResponse{
						Event: installation_entities.PluginInstallEventError,
						Data:  fmt.Sprintf("failed to launch plugin: %s", err.Error()),
					})
					return
				} else {
					responseStream.Write(installation_entities.PluginInstallResponse{
						Event: installation_entities.PluginInstallEventDone,
						Data:  "successfully installed",
					})
					success = true
					return
				}
			}
		}
	})

	return responseStream, nil
}
