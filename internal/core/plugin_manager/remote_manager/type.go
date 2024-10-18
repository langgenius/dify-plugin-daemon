package remote_manager

import (
	"sync"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/basic_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/panjf2000/gnet/v2"
)

type pluginRuntimeMode string

const _PLUGIN_RUNTIME_MODE_CI pluginRuntimeMode = "ci"

type RemotePluginRuntime struct {
	basic_manager.BasicPluginRuntime
	plugin_entities.PluginRuntime

	// connection
	conn   gnet.Conn
	closed int32

	// response entity to accept new events
	response *stream.Stream[[]byte]

	// callbacks for each session
	callbacks      map[string][]func([]byte)
	callbacks_lock *sync.RWMutex

	// channel to notify all waiting routines
	shutdown_chan chan bool

	// heartbeat
	last_active_at time.Time

	// hand shake process completed
	handshake        bool
	handshake_failed bool

	// registration transferred
	registration_transferred bool

	tools_registration_transferred     bool
	models_registration_transferred    bool
	endpoints_registration_transferred bool
	assets_transferred                 bool

	// tenant id
	tenant_id string

	alive bool

	// checksum
	checksum string

	// installation id
	installation_id string

	// wait for started event
	wait_chan_lock          sync.Mutex
	wait_started_chan       []chan bool
	wait_stopped_chan       []chan bool
	wait_launched_chan      chan error
	wait_launched_chan_once sync.Once
}

// Listen creates a new listener for the given session_id
// session id is an unique identifier for a request
func (r *RemotePluginRuntime) addCallback(session_id string, fn func([]byte)) {
	r.callbacks_lock.Lock()
	if _, ok := r.callbacks[session_id]; !ok {
		r.callbacks[session_id] = make([]func([]byte), 0)
	}
	r.callbacks[session_id] = append(r.callbacks[session_id], fn)
	r.callbacks_lock.Unlock()
}

// removeCallback removes the listener for the given session_id
func (r *RemotePluginRuntime) removeCallback(session_id string) {
	r.callbacks_lock.Lock()
	delete(r.callbacks, session_id)
	r.callbacks_lock.Unlock()
}

func (r *RemotePluginRuntime) onDisconnected() {
	// close shutdown channel to notify all waiting routines
	close(r.shutdown_chan)

	// close response to stop current plugin
	r.response.Close()

	r.alive = false
}
