package plugin_manager

import (
	"fmt"
	"sync"
	"time"

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
	// stop plugin when the plugin reaches the end of its lifetime
	defer r.Stop()

	// cleanup plugin runtime state and working directory
	defer r.Cleanup()

	// remove lifetime state after plugin if it has been stopped
	defer r.TriggerStop()

	// register plugin
	for _, reg := range p.pluginRegisters {
		err := reg(r)
		if err != nil {
			log.Error("add plugin to cluster failed: %s", err.Error())
			return
		}
	}

	configuration := r.Configuration()
	log.Info("new plugin logged in: %s", configuration.Identity())
	defer func() {
		log.Info("plugin %s has exited", configuration.Identity())
	}()

	// try to init environment until succeed
	failedTimes := 0

	// only notify launched once
	once := sync.Once{}

	for !r.Stopped() {
		// notify launched if failed too many times
		if failedTimes > 3 {
			once.Do(func() {
				if errChan != nil {
					errChan <- fmt.Errorf(
						"init environment for plugin %s failed too many times, "+
							"you should consider the package is corrupted or your network is unstable",
						configuration.Identity(),
					)
					close(errChan)
				}

				if launchedChan != nil {
					close(launchedChan)
				}
			})

			return
		}

		log.Info("init environment for plugin %s", configuration.Identity())
		if err := r.InitEnvironment(); err != nil {
			if r.Stopped() {
				// plugin has been stopped, exit
				break
			}
			log.Error("init environment failed: %s, retry in 30s", err.Error())
			time.Sleep(30 * time.Second)
			failedTimes++
			continue
		}
		break
	}

	// notify launched
	once.Do(func() {
		if launchedChan != nil {
			close(launchedChan)
		}

		if errChan != nil {
			close(errChan)
		}
	})

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
