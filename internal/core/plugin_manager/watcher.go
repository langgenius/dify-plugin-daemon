package plugin_manager

import (
	"os"
	"path"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/aws_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/local_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

func startWatcher(path string, platform string) {
	go func() {
		log.Info("start to handle new plugins in path: %s", path)
		handleNewPlugins(path, platform)
		for range time.NewTicker(time.Second * 30).C {
			handleNewPlugins(path, platform)
		}
	}()
}

func handleNewPlugins(path string, platform string) {
	// load local plugins firstly
	for plugin := range loadNewPlugins(path) {
		var plugin_interface entities.PluginRuntimeInterface

		if platform == app.PLATFORM_AWS_LAMBDA {
			plugin_interface = &aws_manager.AWSPluginRuntime{
				PluginRuntime: plugin,
			}
		} else if platform == app.PLATFORM_LOCAL {
			plugin_interface = &local_manager.LocalPluginRuntime{
				PluginRuntime: plugin,
			}
		} else {
			log.Error("unsupported platform: %s for plugin: %s", platform, plugin.Config.Name)
			continue
		}

		log.Info("loaded plugin: %s", plugin.Config.Identity())

		m.Store(plugin.Config.Identity(), &plugin)

		routine.Submit(func() {
			lifetime(plugin_interface)
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
				configuration_path := path.Join(root_path, plugin.Name(), "manifest.json")
				configuration, err := parsePluginConfig(configuration_path)
				if err != nil {
					log.Error("parse plugin config error: %v", err)
					continue
				}

				if err := configuration.Validate(); err != nil {
					log.Error("plugin %s config validate error: %v", configuration.Name, err)
					continue
				}

				status := verifyPluginStatus(configuration)
				if status.exist {
					continue
				}

				ch <- entities.PluginRuntime{
					Config: *configuration,
					State: entities.PluginRuntimeState{
						Status:       entities.PLUGIN_RUNTIME_STATUS_PENDING,
						Restarts:     0,
						RelativePath: path.Join(root_path, plugin.Name()),
						ActiveAt:     nil,
						Verified:     false,
					},
				}
			}
		}

		close(ch)
	})

	return ch
}

func parsePluginConfig(configuration_path string) (*entities.PluginConfiguration, error) {
	text, err := os.ReadFile(configuration_path)
	if err != nil {
		return nil, err
	}

	result, err := parser.UnmarshalJson[entities.PluginConfiguration](string(text))
	if err != nil {
		return nil, err
	}

	return &result, nil
}

type pluginStatusResult struct {
	exist bool
}

func verifyPluginStatus(config *entities.PluginConfiguration) pluginStatusResult {
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
