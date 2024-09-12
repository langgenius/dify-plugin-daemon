package remote_manager

import (
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/plugin_errors"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

func (r *RemotePluginRuntime) InitEnvironment() error {
	return nil
}

func (r *RemotePluginRuntime) Stopped() bool {
	return !r.alive
}

func (r *RemotePluginRuntime) Stop() {
	r.alive = false
	if r.conn == nil {
		return
	}
	r.conn.Close()
}

func (r *RemotePluginRuntime) Type() plugin_entities.PluginRuntimeType {
	return plugin_entities.PLUGIN_RUNTIME_TYPE_REMOTE
}

func (r *RemotePluginRuntime) StartPlugin() error {
	var exit_error error

	// handle heartbeat
	routine.Submit(func() {
		r.last_active_at = time.Now()
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if time.Since(r.last_active_at) > 20*time.Second {
					// kill this connection
					r.conn.Close()
					exit_error = plugin_errors.ErrPluginNotActive
					return
				}
			case <-r.shutdown_chan:
				return
			}
		}
	})

	r.response.Async(func(data []byte) {
		plugin_entities.ParsePluginUniversalEvent(
			data,
			func(session_id string, data []byte) {
				r.callbacks_lock.RLock()
				listeners := r.callbacks[session_id][:]
				r.callbacks_lock.RUnlock()

				// handle session event
				for _, listener := range listeners {
					listener(data)
				}
			},
			func() {
				r.last_active_at = time.Now()
			},
			func(err string) {
				log.Error("plugin %s: %s", r.Configuration().Identity(), err)
			},
			func(message string) {
				log.Info("plugin %s: %s", r.Configuration().Identity(), message)
			},
		)
	})

	return exit_error
}

func (r *RemotePluginRuntime) Wait() (<-chan bool, error) {
	return r.shutdown_chan, nil
}

func (r *RemotePluginRuntime) Checksum() (string, error) {
	return r.checksum, nil
}
