package plugin_manager

import (
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

func (p *PluginManager) AddPluginRegisterHandler(handler func(r plugin_entities.PluginLifetime) error) {
	p.pluginRegisters = append(p.pluginRegisters, handler)
}

func (p *PluginManager) fullDuplexLifetime(r plugin_entities.PluginFullDuplexLifetime) {
	configuration := r.Configuration()

	log.Info("new plugin logged in: %s", configuration.Identity())
	defer log.Info("plugin %s has exited", configuration.Identity())

	// cleanup plugin runtime state and working directory
	defer r.Cleanup()

	// stop plugin when the plugin reaches the end of its lifetime
	defer r.Stop()

	// register plugin
	for _, reg := range p.pluginRegisters {
		err := reg(r)
		if err != nil {
			log.Error("add plugin to cluster failed: %s", err.Error())
			return
		}
	}

	start_failed_times := 0

	// remove lifetime state after plugin if it has been stopped
	defer r.TriggerStop()

	// try at most 3 times to init environment
	for i := 0; i < 3; i++ {
		if err := r.InitEnvironment(); err != nil {
			log.Error("init environment failed: %s, retry in 30s", err.Error())
			if start_failed_times == 3 {
				log.Error(
					"init environment failed 3 times, plugin %s has been stopped",
					configuration.Identity(),
				)
				return
			}
			time.Sleep(30 * time.Second)
			start_failed_times++
			continue
		}
	}

	// TODO: launched will only be triggered when calling StartPlugin

	// init environment successfully
	// once succeed, we consider the plugin is installed successfully
	for !r.Stopped() {
		// start plugin
		if err := r.StartPlugin(); err != nil {
			if r.Stopped() {
				// plugin has been stopped, exit
				break
			}
		}

		// wait for plugin to stop normally
		c, err := r.Wait()
		if err == nil {
			<-c
		}

		// restart plugin in 5s
		time.Sleep(5 * time.Second)

		// add restart times
		r.AddRestarts()
	}
}
