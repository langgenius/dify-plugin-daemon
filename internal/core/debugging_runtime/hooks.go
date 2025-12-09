package debugging_runtime

import (
	"bytes"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/basic_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/media_transport"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/stream"
	"github.com/panjf2000/gnet/v2"
)

type DifyServer struct {
	gnet.BuiltinEventEngine

	engine gnet.Engine

	mediaManager *media_transport.MediaBucket

	// listening address
	addr string
	port uint16

	// enabled multicore
	multicore bool

	// event loop count
	numLoops int

	plugins     map[int]*RemotePluginRuntime
	pluginsLock *sync.RWMutex

	maxConn     int32
	currentConn int32

	notifiers     []PluginRuntimeNotifier
	notifierMutex *sync.RWMutex
}

func (s *DifyServer) OnBoot(c gnet.Engine) (action gnet.Action) {
	s.engine = c
	return gnet.None
}

func (s *DifyServer) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	// new plugin connected
	c.SetContext(&codec{})
	runtime := &RemotePluginRuntime{
		MediaTransport: basic_runtime.NewMediaTransport(
			s.mediaManager,
		),

		conn:                      c,
		response:                  stream.NewStream[[]byte](512),
		messageCallbacks:          make(map[string][]func([]byte)),
		messageCallbacksLock:      &sync.RWMutex{},
		sessionMessageClosers:     make(map[string][]func()),
		sessionMessageClosersLock: &sync.RWMutex{},

		assets:      make(map[string]*bytes.Buffer),
		assetsBytes: 0,
		alive:       true,
	}

	// store plugin runtime
	s.pluginsLock.Lock()
	s.plugins[c.Fd()] = runtime
	s.pluginsLock.Unlock()

	// start a timer to check if handshake is completed in 10 seconds
	time.AfterFunc(time.Second*10, func() {
		if !runtime.handshake {
			// close connection
			c.Close()
		}
	})

	// verified
	verified := true
	if verified {
		return nil, gnet.None
	}

	return nil, gnet.Close
}

func (s *DifyServer) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	// plugin disconnected
	s.pluginsLock.Lock()
	plugin := s.plugins[c.Fd()]
	delete(s.plugins, c.Fd())
	s.pluginsLock.Unlock()

	if plugin == nil {
		return gnet.None
	}

	// close plugin
	plugin.cleanupResources()

	// trigger runtime disconnected event
	s.WalkNotifiers(func(notifier PluginRuntimeNotifier) {
		notifier.OnRuntimeDisconnected(plugin)
	})

	// decrease current connection
	atomic.AddInt32(&s.currentConn, -1)

	return gnet.None
}

func (s *DifyServer) OnShutdown(c gnet.Engine) {
	s.WalkNotifiers(func(notifier PluginRuntimeNotifier) {
		notifier.OnServerShutdown(SERVER_SHUTDOWN_REASON_EXIT)
	})
}

func (s *DifyServer) OnTraffic(c gnet.Conn) (action gnet.Action) {
	defer func() {
		if r := recover(); r != nil {
			traceback := string(debug.Stack())
			log.Error("panic in OnTraffic: %v\n%s", r, traceback)
		}
	}()

	codec := c.Context().(*codec)
	messages, err := codec.Decode(c)
	if err != nil {
		return gnet.Close
	}

	// get plugin runtime
	s.pluginsLock.RLock()
	runtime, ok := s.plugins[c.Fd()]
	s.pluginsLock.RUnlock()
	if !ok {
		return gnet.Close
	}

	// handle messages
	for _, message := range messages {
		if len(message) == 0 {
			continue
		}

		s.onMessage(runtime, message)
	}

	return gnet.None
}

func (s *DifyServer) onMessage(runtime *RemotePluginRuntime, message []byte) {
	// handle message
	if runtime.handshakeFailed {
		// do nothing if handshake has failed
		return
	}

	closeConn := func(message []byte) {
		if atomic.CompareAndSwapInt32(&runtime.closed, 0, 1) {
			runtime.conn.Write(message)
			runtime.conn.Close()
		}
	}

	if !runtime.initialized {
		registerPayload, err := parser.UnmarshalJsonBytes[plugin_entities.RemotePluginRegisterPayload](message)
		if err != nil {
			// close connection if handshake failed
			closeConn([]byte("handshake failed, invalid handshake message\n"))
			runtime.handshakeFailed = true
			return
		}

		switch registerPayload.Type {
		case plugin_entities.REGISTER_EVENT_TYPE_HAND_SHAKE:
			if connectionInfo, err := s.handleHandleShake(runtime, registerPayload); err != nil {
				runtime.handshakeFailed = true
				closeConn(append([]byte(err.Error()), '\n'))
			} else {
				runtime.tenantId = connectionInfo.TenantId
				runtime.handshake = true
			}
		case plugin_entities.REGISTER_EVENT_TYPE_ASSET_CHUNK:
			if err := s.handleAssetsTransfer(runtime, registerPayload); err != nil {
				closeConn(append([]byte(err.Error()), '\n'))
			}
		case plugin_entities.REGISTER_EVENT_TYPE_END:
			atomic.AddInt32(&s.currentConn, 1)
			if atomic.LoadInt32(&s.currentConn) > int32(s.maxConn) {
				closeConn([]byte("server is busy now, please try again later\n"))
				return
			}
			if err := s.handleInitializationEndEvent(runtime); err != nil {
				closeConn(append([]byte(err.Error()), '\n'))
				return
			}

			// trigger new connection event
			s.WalkNotifiers(func(notifier PluginRuntimeNotifier) {
				notifier.OnRuntimeConnected(runtime)
			})
		case plugin_entities.REGISTER_EVENT_TYPE_MANIFEST_DECLARATION:
			if err := s.handleDeclarationRegister(runtime, registerPayload); err != nil {
				closeConn(append([]byte(err.Error()), '\n'))
			}
		case plugin_entities.REGISTER_EVENT_TYPE_TOOL_DECLARATION:
			if err := s.handleToolDeclarationRegister(runtime, registerPayload); err != nil {
				closeConn(append([]byte(err.Error()), '\n'))
			}
		case plugin_entities.REGISTER_EVENT_TYPE_MODEL_DECLARATION:
			if err := s.handleModelDeclarationRegister(runtime, registerPayload); err != nil {
				closeConn(append([]byte(err.Error()), '\n'))
			}
		case plugin_entities.REGISTER_EVENT_TYPE_ENDPOINT_DECLARATION:
			if err := s.handleEndpointDeclarationRegister(runtime, registerPayload); err != nil {
				closeConn(append([]byte(err.Error()), '\n'))
			}
		case plugin_entities.REGISTER_EVENT_TYPE_AGENT_STRATEGY_DECLARATION:
			if err := s.handleAgentStrategyDeclarationRegister(runtime, registerPayload); err != nil {
				closeConn(append([]byte(err.Error()), '\n'))
			}
		case plugin_entities.REGISTER_EVENT_TYPE_DATASOURCE_DECLARATION:
			if err := s.handleDatasourceDeclarationRegister(runtime, registerPayload); err != nil {
				closeConn(append([]byte(err.Error()), '\n'))
			}
		case plugin_entities.REGISTER_EVENT_TYPE_TRIGGER_DECLARATION:
			if err := s.handleTriggerDeclarationRegister(runtime, registerPayload); err != nil {
				closeConn(append([]byte(err.Error()), '\n'))
			}
		}
	} else {
		// continue handle messages if handshake completed
		runtime.response.WriteBlocking(message)
	}
}

// AddNotifier adds a notifier to the runtime
func (r *DifyServer) AddNotifier(notifier PluginRuntimeNotifier) {
	r.notifierMutex.Lock()
	defer r.notifierMutex.Unlock()

	r.notifiers = append(r.notifiers, notifier)
}

// WalkNotifiers walks through all the notifiers and calls the given function
func (r *DifyServer) WalkNotifiers(fn func(notifier PluginRuntimeNotifier)) {
	r.notifierMutex.RLock()
	notifiers := r.notifiers // copy the notifiers
	r.notifierMutex.RUnlock()

	for _, notifier := range notifiers {
		fn(notifier)
	}
}
