package plugin_manager

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/lifecycle"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func (p *PluginManager) AddPluginRegisterHandler(handler func(r plugin_entities.PluginLifetime) error) {
	p.pluginRegisters = append(p.pluginRegisters, handler)
}

// fullDuplexLifecycle takes the responsibility of full-duplex lifecycle of a plugin
// it will block the thread until the plugin is stopped so it's important to call it in a new goroutine
//  1. try to init environment until succeed or plugin has failed too many times
//  2. launchedChan and errChan are used to synchronize the plugin launch process
//     only if received non-nil message from errChan, it's considered the setup process has failed
//  3. after exit, environment will be cleaned up
//
// NOTE: the size of launchedChan and errChan should always be 0 to keep the sync mechanism working
func (p *PluginManager) fullDuplexLifecycle(
	r plugin_entities.PluginFullDuplexLifetime,
	launchedChan chan bool,
	errChan chan error,
) {

	// register plugin
	for _, reg := range p.pluginRegisters {
		err := reg(r)
		if err != nil {
			log.Error("add plugin to cluster failed: %s", err.Error())
			return
		}
	}

	lifecycle.FullDuplex(r, launchedChan, errChan)
}
