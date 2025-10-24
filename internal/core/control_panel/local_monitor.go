package controlpanel

import (
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

func (c *ControlPanel) StartLocalMonitor() {
	log.Info("start to handle new plugins in path: %s", c.config.PluginInstalledPath)
	log.Info("launch plugins with max concurrency: %d", c.config.PluginLocalLaunchingConcurrent)

	c.handleNewLocalPlugins()
	// sync every 30 seconds
	for range time.NewTicker(time.Second * 30).C {
		c.handleNewLocalPlugins()
	}
}

func (c *ControlPanel) handleNewLocalPlugins() {
	// TODO: handle new local plugins
}
