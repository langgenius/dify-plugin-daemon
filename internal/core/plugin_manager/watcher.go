package plugin_manager

import (
	"os"
	"path"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/aws_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/local_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/remote_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

func startWatcher(config *app.Config) {
	go func() {
		log.Info("start to handle new plugins in path: %s", config.StoragePath)
		handleNewPlugins(config)
		for range time.NewTicker(time.Second * 30).C {
			handleNewPlugins(config)
		}
	}()

	startRemoteWatcher(config)
}

func startRemoteWatcher(config *app.Config) {
	// launch TCP debugging server if enabled
	if config.PluginRemoteInstallingEnabled {
		server := remote_manager.NewRemotePluginServer(config)
		go func() {
			err := server.Launch()
			if err != nil {
				log.Error("start remote plugin server failed: %s", err.Error())
			}
		}()
		go func() {
			server.Wrap(func(rpr *remote_manager.RemotePluginRuntime) {
				lifetime(config, rpr)
			})
		}()
	}
}

func handleNewPlugins(config *app.Config) {
	// load local plugins firstly
	for plugin := range loadNewPlugins(config.StoragePath) {
		var plugin_interface entities.PluginRuntimeInterface

		if config.Platform == app.PLATFORM_AWS_LAMBDA {
			plugin_interface = &aws_manager.AWSPluginRuntime{
				PluginRuntime: plugin,
			}
		} else if config.Platform == app.PLATFORM_LOCAL {
			plugin_interface = &local_manager.LocalPluginRuntime{
				PluginRuntime: plugin,
			}
		} else {
			log.Error("unsupported platform: %s for plugin: %s", config.Platform, plugin.Config.Name)
			continue
		}

		routine.Submit(func() {
			lifetime(config, plugin_interface)
		})
	}
}

// chan should be closed after using that
func loadNewPlugins(root_path string) <-chan entities.PluginRuntime {
	ch := make(chan entities.PluginRuntime)

	plugins, err := os.ReadDir(root_path)
	if err != nil {
		log.Error("no plugin found in path: %s", root_path)
		close(ch)
		return ch
	}

	routine.Submit(func() {
		for _, plugin := range plugins {
			if plugin.IsDir() {
				configuration_path := path.Join(root_path, plugin.Name(), "manifest.yaml")
				configuration, err := parsePluginConfig(configuration_path)
				if err != nil {
					log.Error("parse plugin config error: %v", err)
					continue
				}

				status := verifyPluginStatus(configuration)
				if status.exist {
					continue
				}

				// check if .verified file exists
				verified_path := path.Join(root_path, plugin.Name(), ".verified")
				_, err = os.Stat(verified_path)

				ch <- entities.PluginRuntime{
					Config: *configuration,
					State: entities.PluginRuntimeState{
						Status:       entities.PLUGIN_RUNTIME_STATUS_PENDING,
						Restarts:     0,
						RelativePath: path.Join(root_path, plugin.Name()),
						ActiveAt:     nil,
						Verified:     err == nil,
					},
				}
			}
		}

		close(ch)
	})

	return ch
}

func parsePluginConfig(configuration_path string) (*plugin_entities.PluginDeclaration, error) {
	text, err := os.ReadFile(configuration_path)
	if err != nil {
		return nil, err
	}

	result, err := plugin_entities.UnmarshalPluginDeclarationFromYaml(text)
	if err != nil {
		return nil, err
	}

	return result, nil
}

type pluginStatusResult struct {
	exist bool
}

func verifyPluginStatus(config *plugin_entities.PluginDeclaration) pluginStatusResult {
	_, exist := checkPluginExist(config.Identity())
	if exist {
		return pluginStatusResult{
			exist: true,
		}
	}

	return pluginStatusResult{
		exist: false,
	}
}
