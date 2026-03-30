package debugging_runtime

import (
	"time"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

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

// spawn a core to handle CPU-intensive tasks
func (r *RemotePluginRuntime) SpawnCore() error {
	var exitError error

	r.response.Process(func(data []byte) {
		plugin_entities.ParsePluginUniversalEvent(
			data,
			"",
			func(session_id string, data []byte) {
				r.messageCallbacksLock.RLock()
				listeners := r.messageCallbacks[session_id][:]
				r.messageCallbacksLock.RUnlock()

				// handle session event
				for _, listener := range listeners {
					listener(data)
				}
			},
			func() {
				r.lastActiveAt = time.Now()
			},
			func(err string) {
				log.Error("plugin error", "plugin", r.Configuration().Identity(), "error", err)
			},
			func(plugin_entities.PluginLogEvent) {}, // remote debugging logs stay on client side
		)
	})

	return exitError
}

func (r *RemotePluginRuntime) HeartbeatMonitor() {
	// close connection if it hangs for over 60 seconds
	ticker := time.NewTicker(time.Second * 60)
	defer ticker.Stop()

	for range ticker.C {
		if r.Stopped() {
			return
		}

		if time.Since(r.lastActiveAt) > time.Second*60 {
			r.Stop()
		}
	}
}

func (r *RemotePluginRuntime) Checksum() (string, error) {
	return r.checksum, nil
}
