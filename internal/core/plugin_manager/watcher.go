package plugin_manager

import (
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/debugging_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func (p *PluginManager) startLocalWatcher() {
	go func() {
		log.Info("start to handle new plugins in path: %s", p.config.PluginInstalledPath)
		p.handleNewLocalPlugins()
		for range time.NewTicker(time.Second * 30).C {
			p.handleNewLocalPlugins()
			p.removeUninstalledLocalPlugins()
		}
	}()
}

func (p *PluginManager) initRemotePluginServer(config *app.Config) {
	if p.remotePluginServer != nil {
		return
	}
	p.remotePluginServer = debugging_runtime.NewRemotePluginServer(config, p.mediaBucket)
}

func (p *PluginManager) startRemoteWatcher(config *app.Config) {
	// launch TCP debugging server if enabled
	if config.PluginRemoteInstallingEnabled != nil && *config.PluginRemoteInstallingEnabled {
		p.initRemotePluginServer(config)
		go func() {
			err := p.remotePluginServer.Launch()
			if err != nil {
				log.Error("start remote plugin server failed: %s", err.Error())
			}
		}()
		go func() {
			p.remotePluginServer.Wrap(func(rpr plugin_entities.PluginFullDuplexLifetime) {
				identity, err := rpr.Identity()
				if err != nil {
					log.Error("get remote plugin identity failed: %s", err.Error())
					return
				}
				p.m.Store(identity.String(), rpr)
				routine.Submit(map[string]string{
					"module":    "plugin_manager",
					"function":  "startRemoteWatcher",
					"plugin_id": identity.String(),
					"type":      "remote",
				}, func() {
					defer func() {
						if err := recover(); err != nil {
							log.Error("plugin runtime error: %v", err)
						}
						p.m.Delete(identity.String())
					}()
					p.fullDuplexLifecycle(rpr, nil, nil)
				})
			})
		}()
	}
}

func (p *PluginManager) handleNewLocalPlugins() {
	// walk through all plugins
	plugins, err := p.installedBucket.List()
	if err != nil {
		log.Error("list installed plugins failed: %s", err.Error())
		return
	}

	for _, plugin := range plugins {
		_, launchedChan, errChan, err := p.launchLocal(plugin)
		if err != nil {
			log.Error("launch local plugin failed: %s", err.Error())
		}

		// avoid receiving nil channel
		if errChan != nil {
			// consume error, avoid deadlock
			for err := range errChan {
				log.Error("plugin launch error: %s", err.Error())
			}
		}

		// avoid receiving nil channel
		if launchedChan != nil {
			// wait for plugin launched
			<-launchedChan
		}
	}
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
