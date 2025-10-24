package plugin_manager

import (
	"sync"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func (p *PluginManager) startLocalWatcher(config *app.Config) {
	// REMOVE
}

func (p *PluginManager) initRemotePluginServer(config *app.Config) {
	// REMOVE
}

func (p *PluginManager) startRemoteWatcher(config *app.Config) {
	// REMOVE
}

func (p *PluginManager) handleNewLocalPlugins(config *app.Config) {
	// walk through all plugins
	plugins, err := p.installedBucket.List()
	if err != nil {
		log.Error("list installed plugins failed: %s", err.Error())
		return
	}

	var wg sync.WaitGroup
	maxConcurrency := config.PluginLocalLaunchingConcurrent
	sem := make(chan struct{}, maxConcurrency)

	for _, plugin := range plugins {
		_, exist := p.m.Load(plugin.String())
		if exist {
			continue
		}

		wg.Add(1)
		// Fix closure issue: create local variable copy
		currentPlugin := plugin
		routine.Submit(map[string]string{
			"module":   "plugin_manager",
			"function": "handleNewLocalPlugins",
		}, func() {
			// Acquire sem inside goroutine
			sem <- struct{}{}
			defer func() {
				if err := recover(); err != nil {
					log.Error("plugin launch runtime error: %v", err)
				}
				<-sem
				wg.Done()
			}()

			_, launchedChan, errChan, err := p.launchLocal(currentPlugin)
			if err != nil {
				log.Error("launch local plugin failed: %s", err.Error())
				return
			}

			// Handle error channel
			if errChan != nil {
				for err := range errChan {
					log.Error("plugin launch error: %s", err.Error())
				}
			}

			// Wait for plugin to complete startup
			if launchedChan != nil {
				<-launchedChan
			}
		})
	}

	// wait for all plugins to be launched
	wg.Wait()
}

// an async function to remove uninstalled local plugins
func (p *PluginManager) removeUninstalledLocalPlugins() {
	// read all local plugin runtimes
	p.m.Range(func(key string, value plugin_entities.PluginLifetime) bool {
		// try to convert to local runtime
		runtime, ok := value.(*local_runtime.LocalPluginRuntime)
		if !ok {
			return true
		}

		pluginUniqueIdentifier, err := runtime.Identity()
		if err != nil {
			log.Error("get plugin identity failed: %s", err.Error())
			return true
		}

		// check if plugin is deleted, stop it if so
		exists, err := p.installedBucket.Exists(pluginUniqueIdentifier)
		if err != nil {
			log.Error("check if plugin is deleted failed: %s", err.Error())
			return true
		}

		if !exists {
			runtime.Stop()
		}

		return true
	})
}
