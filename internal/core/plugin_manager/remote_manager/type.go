package remote_manager

import (
	"bytes"
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
	callbacks     map[string][]func([]byte)
	callbacksLock *sync.RWMutex

	// channel to notify all waiting routines
	shutdownChan chan bool

	// heartbeat
	lastActiveAt time.Time

	assets      map[string]*bytes.Buffer
	assetsBytes int64

	// hand shake process completed
	handshake       bool
	handshakeFailed bool

	// initialized, wether registration transferred
	initialized bool

	// registration transferred
	registrationTransferred bool

	toolsRegistrationTransferred     bool
	modelsRegistrationTransferred    bool
	endpointsRegistrationTransferred bool
	agentsRegistrationTransferred    bool
	assetsTransferred                bool

	// tenant id
	tenantId string

	alive bool

	// checksum
	checksum string

	// installation id
	installationId string

	// wait for started event
	waitChanLock         sync.Mutex
	waitStartedChan      []chan bool
	waitStoppedChan      []chan bool
	waitLaunchedChan     chan error
	waitLaunchedChanOnce sync.Once
}

// Listen creates a new listener for the given session_id
// session id is an unique identifier for a request
func (r *RemotePluginRuntime) addCallback(session_id string, fn func([]byte)) {
	r.callbacksLock.Lock()
	if _, ok := r.callbacks[session_id]; !ok {
		r.callbacks[session_id] = make([]func([]byte), 0)
	}
	r.callbacks[session_id] = append(r.callbacks[session_id], fn)
	r.callbacksLock.Unlock()
}

// removeCallback removes the listener for the given session_id
func (r *RemotePluginRuntime) removeCallback(session_id string) {
	r.callbacksLock.Lock()
	delete(r.callbacks, session_id)
	r.callbacksLock.Unlock()
}

func (r *RemotePluginRuntime) onDisconnected() {
	// close shutdown channel to notify all waiting routines
	close(r.shutdownChan)

	// close response to stop current plugin
	r.response.Close()

	r.alive = false
}
