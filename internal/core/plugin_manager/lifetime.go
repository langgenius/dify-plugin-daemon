package plugin_manager

import (
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

func lifetime(r entities.PluginRuntimeInterface) {
	start_failed_times := 0

	for !r.Stopped() {
		if err := r.InitEnvironment(); err != nil {
			log.Error("init environment failed: %s, retry in 30s", err.Error())
			time.Sleep(30 * time.Second)
			if start_failed_times == 3 {
				log.Error(
					"init environment failed 3 times, plugin %s has been stopped",
					r.Configuration().Identity(),
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
					r.Configuration().Identity(),
				)
				r.Stop()
			}

			start_failed_times++
			continue
		}
	}
}
