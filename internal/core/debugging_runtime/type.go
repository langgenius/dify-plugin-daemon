package debugging_runtime

import (
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/basic_runtime"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/stream"
	"github.com/panjf2000/gnet/v2"
)

type RemotePluginRuntime struct {
	basic_runtime.MediaTransport
	plugin_entities.PluginRuntime

	// connection
	conn   gnet.Conn
	closed int32

	// response entity to accept new events
	response *stream.Stream[[]byte]

	// messageCallbacks for each session
	messageCallbacks     map[string][]func([]byte)
	messageCallbacksLock *sync.RWMutex

	// sessionMessageCloser for each session
	sessionMessageClosers     map[string][]func()
	sessionMessageClosersLock *sync.RWMutex

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

	toolsRegistrationTransferred         bool
	modelsRegistrationTransferred        bool
	endpointsRegistrationTransferred     bool
	agentStrategyRegistrationTransferred bool
	datasourceRegistrationTransferred    bool
	triggersRegistrationTransferred      bool
	assetsTransferred                    bool

	// tenant id
	tenantId string

	alive bool

	// checksum
	checksum string

	// installation id
	installationId string

	// wait for started event
	waitChanLock    sync.Mutex
	waitStartedChan []chan bool
	waitStoppedChan []chan bool
}

// Listen creates a new listener for the given session_id
// session id is an unique identifier for a request
func (r *RemotePluginRuntime) addMessageCallbackHandler(session_id string, fn func([]byte)) {
	r.messageCallbacksLock.Lock()
	if _, ok := r.messageCallbacks[session_id]; !ok {
		r.messageCallbacks[session_id] = make([]func([]byte), 0)
	}
	r.messageCallbacks[session_id] = append(r.messageCallbacks[session_id], fn)
	r.messageCallbacksLock.Unlock()
}

// removeMessageCallbackHandler removes the listener for the given session_id
func (r *RemotePluginRuntime) removeMessageCallbackHandler(session_id string) {
	r.messageCallbacksLock.Lock()
	delete(r.messageCallbacks, session_id)
	r.messageCallbacksLock.Unlock()
}

// addSessionMessageCloser adds a closer for the given session_id
// once the session is closed or the connection is closed, the closer will be called
func (r *RemotePluginRuntime) addSessionMessageCloser(session_id string, fn func()) {
	// do nothing if the session is already closed
	if atomic.LoadInt32(&r.closed) == 1 {
		return
	}

	r.sessionMessageClosersLock.Lock()
	if _, ok := r.sessionMessageClosers[session_id]; !ok {
		r.sessionMessageClosers[session_id] = make([]func(), 0)
	}
	r.sessionMessageClosers[session_id] = append(r.sessionMessageClosers[session_id], fn)
	r.sessionMessageClosersLock.Unlock()
}

// removeSessionMessageCloser removes the closer for the given session_id
func (r *RemotePluginRuntime) removeSessionMessageCloser(session_id string) {
	// do nothing if the session is already closed
	if atomic.LoadInt32(&r.closed) == 1 {
		return
	}

	r.sessionMessageClosersLock.Lock()
	delete(r.sessionMessageClosers, session_id)
	r.sessionMessageClosersLock.Unlock()
}

func (r *RemotePluginRuntime) cleanupResources() {
	// call all session message closers
	r.sessionMessageClosersLock.RLock()
	for _, closer := range r.sessionMessageClosers {
		for _, fn := range closer {
			fn()
		}
	}
	r.sessionMessageClosersLock.RUnlock()

	// change the alive status
	r.alive = false

	// change the closed status
	atomic.StoreInt32(&r.closed, 1)

	// close response to stop current plugin
	r.response.Close()

	r.alive = false
}

func (r *RemotePluginRuntime) TenantId() string {
	return r.tenantId
}

func (r *RemotePluginRuntime) InstallationId() string {
	return r.installationId
}

func (r *RemotePluginRuntime) SetInstallationId(installationId string) {
	// FIXME(Yeuoly): temporary solution for managing plugin installation model in DB
	// once a connection is disconnected, removing it needs installation id
	// however, it was set outside of the plugin runtime, it uses `notifiers`
	r.installationId = installationId
}

func (r *RemotePluginRuntime) Identity() (plugin_entities.PluginUniqueIdentifier, error) {
	// FIXME: it's a little bit tricky that replace author with current tenant_id
	// just as a flag to identify debugging plugin
	config := r.Config
	config.Author = r.tenantId
	checksum, _ := r.Checksum()
	return plugin_entities.NewPluginUniqueIdentifier(fmt.Sprintf("%s@%s", config.Identity(), checksum))
}
