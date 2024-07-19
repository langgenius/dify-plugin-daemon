package plugin_manager

import (
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

func lifetime(config *app.Config, r entities.PluginRuntimeInterface) {
	start_failed_times := 0
	configuration := r.Configuration()

	addLifetimeState(r)
	defer func() {
		// remove lifetime state after plugin if it has been stopped for $LIFETIME_STATE_GC_INTERVAL and not started again
		time.AfterFunc(time.Duration(config.LifetimeStateGCInterval)*time.Second, func() {
			if r.Stopped() {
				deleteLifetimeState(r)
			}
		})
	}()

	for !r.Stopped() {
		if err := r.InitEnvironment(); err != nil {
			log.Error("init environment failed: %s, retry in 30s", err.Error())
			time.Sleep(30 * time.Second)
			if start_failed_times == 3 {
				log.Error(
					"init environment failed 3 times, plugin %s has been stopped",
					configuration.Identity(),
				)
				r.Stop()
			}
			start_failed_times++
			continue
		}

		start_failed_times = 0
		// start plugin
		if err := r.StartPlugin(); err != nil {
			log.Error("start plugin failed: %s, retry in 30s", err.Error())
			time.Sleep(30 * time.Second)
			if start_failed_times == 3 {
				log.Error(
					"start plugin failed 3 times, plugin %s has been stopped",
					configuration.Identity(),
				)
				r.Stop()
			}

			start_failed_times++
			continue
		}

		// wait for plugin to stop
		c, err := r.Wait()
		if err == nil {
			<-c
		}

		// restart plugin in 5s
		time.Sleep(5 * time.Second)
	}
}
