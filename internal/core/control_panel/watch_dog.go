package controlpanel

import (
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

func (c *ControlPanel) StartWatchDog() {
	// start local plugin watch dog
	c.startLocalPluginWatchDog()

	// start debugging server watch dog
	c.startDebuggingServerWatchDog()
}

func (c *ControlPanel) startLocalPluginWatchDog() {
	if c.config.Platform == app.PLATFORM_LOCAL {
		// start local monitor
		go c.startLocalMonitor()

		// continue check if a plugin was uninstalled
		// AS plugin_daemon supports cluster mode
		// installed plugins were stored in `c.installedBucket`
		// it's a stateless across all plugin_daemon nodes
		// a plugin may be uninstalled by other nodes
		// to ensure all uninstalled plugin running in all nodes are stopped
		go c.removeUnusedLocalPlugins()
	}
}

func (c *ControlPanel) startDebuggingServerWatchDog() {
	// launch TCP debugging server if enabled
	if c.config.PluginRemoteInstallingEnabled {
		// setup debugging server
		c.setupDebuggingServer(c.config)

		// launch debugging server
		go func() {
			err := c.startDebuggingServer()
			if err != nil {
				log.Error("start remote plugin server failed: %s", err.Error())
			}
		}()
	}
}

// The plugin can never be launched by `WatchDog` automatically
//
// by default, all plugins are in auto launch mode
func (c *ControlPanel) DisableLocalPluginAutoLaunch(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) {
	c.localPluginWatchIgnoreList.Store(pluginUniqueIdentifier, true)
}

// Enable auto launch for a plugin
//
// by default, all plugins are in auto launch mode
func (c *ControlPanel) EnableLocalPluginAutoLaunch(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) {
	c.localPluginWatchIgnoreList.Delete(pluginUniqueIdentifier)
}
