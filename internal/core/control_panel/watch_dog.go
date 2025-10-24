package controlpanel

import (
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

func (c *ControlPanel) StartWatchDog() {
	// start local plugin watch dog
	c.StartLocalPluginWatchDog()

	// start debugging server watch dog
	c.StartDebuggingServerWatchDog()
}

func (c *ControlPanel) StartLocalPluginWatchDog() {
	// start local monitor
	go c.StartLocalMonitor()
}

func (c *ControlPanel) StartDebuggingServerWatchDog() {
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

		// consume debugging plugin
		go func() {
			c.consumeDebuggingPlugin()
		}()
	}
}
