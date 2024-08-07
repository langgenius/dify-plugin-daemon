package remote_manager

import (
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/plugin_errors"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
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

func (r *RemotePluginRuntime) Type() entities.PluginRuntimeType {
	return entities.PLUGIN_RUNTIME_TYPE_REMOTE
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

	r.response.Wrap(func(data []byte) {
		// handle event
		event, err := parser.UnmarshalJsonBytes[plugin_entities.PluginUniversalEvent](data)
		if err != nil {
			return
		}

		session_id := event.SessionId

		switch event.Event {
		case plugin_entities.PLUGIN_EVENT_LOG:
			if event.Event == plugin_entities.PLUGIN_EVENT_LOG {
				log_event, err := parser.UnmarshalJsonBytes[plugin_entities.PluginLogEvent](
					event.Data,
				)
				if err != nil {
					log.Error("unmarshal json failed: %s", err.Error())
					return
				}

				log.Info("plugin %s: %s", r.Configuration().Identity(), log_event.Message)
			}
		case plugin_entities.PLUGIN_EVENT_SESSION:
			r.callbacks_lock.RLock()
			listeners := r.callbacks[session_id][:]
			r.callbacks_lock.RUnlock()

			// handle session event
			for _, listener := range listeners {
				listener(event.Data)
			}
		case plugin_entities.PLUGIN_EVENT_ERROR:
			log.Error("plugin %s: %s", r.Configuration().Identity(), event.Data)
		case plugin_entities.PLUGIN_EVENT_HEARTBEAT:
			r.last_active_at = time.Now()
		}
	})

	return exit_error
}

func (r *RemotePluginRuntime) Wait() (<-chan bool, error) {
	return r.shutdown_chan, nil
}

func (r *RemotePluginRuntime) Checksum() string {
	return r.checksum
}
