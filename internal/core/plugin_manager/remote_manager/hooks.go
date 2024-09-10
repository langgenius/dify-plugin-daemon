package remote_manager

import (
	"encoding/hex"
	"sync"
	"sync/atomic"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/basic_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/media_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/panjf2000/gnet/v2"
)

var (
	_mode pluginRuntimeMode
)

type DifyServer struct {
	gnet.BuiltinEventEngine

	engine gnet.Engine

	mediaManager *media_manager.MediaManager

	// listening address
	addr string
	port uint16

	// enabled multicore
	multicore bool

	// event loop count
	num_loops int

	// read new connections
	response *stream.Stream[*RemotePluginRuntime]

	plugins      map[int]*RemotePluginRuntime
	plugins_lock *sync.RWMutex

	shutdown_chan chan bool
}

func (s *DifyServer) OnBoot(c gnet.Engine) (action gnet.Action) {
	s.engine = c
	return gnet.None
}

func (s *DifyServer) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	// new plugin connected
	c.SetContext(&codec{})
	runtime := &RemotePluginRuntime{
		BasicPluginRuntime: basic_manager.NewBasicPluginRuntime(
			s.mediaManager,
		),

		conn:           c,
		response:       stream.NewStreamResponse[[]byte](512),
		callbacks:      make(map[string][]func([]byte)),
		callbacks_lock: &sync.RWMutex{},

		shutdown_chan: make(chan bool),

		alive: true,
	}

	// store plugin runtime
	s.plugins_lock.Lock()
	s.plugins[c.Fd()] = runtime
	s.plugins_lock.Unlock()

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
	s.plugins_lock.Lock()
	plugin := s.plugins[c.Fd()]
	delete(s.plugins, c.Fd())
	s.plugins_lock.Unlock()

	// close plugin
	plugin.onDisconnected()

	// clear assets
	plugin.ClearAssets()

	// uninstall plugin
	if plugin.assets_transferred {
		if _mode != _PLUGIN_RUNTIME_MODE_CI {
			if err := plugin.Unregister(); err != nil {
				log.Error("unregister plugin failed, error: %v", err)
			}
		}
	}

	return gnet.None
}

func (s *DifyServer) OnShutdown(c gnet.Engine) {
	close(s.shutdown_chan)
}

func (s *DifyServer) OnTraffic(c gnet.Conn) (action gnet.Action) {
	codec := c.Context().(*codec)
	messages, err := codec.Decode(c)
	if err != nil {
		return gnet.Close
	}

	// get plugin runtime
	s.plugins_lock.RLock()
	runtime, ok := s.plugins[c.Fd()]
	s.plugins_lock.RUnlock()
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
	if runtime.handshake_failed {
		// do nothing if handshake has failed
		return
	}

	close := func(message []byte) {
		if atomic.CompareAndSwapInt32(&runtime.closed, 0, 1) {
			runtime.conn.Write(message)
			runtime.conn.Close()
		}
	}

	if !runtime.handshake {
		key := string(message)

		info, err := GetConnectionInfo(key)
		if err == cache.ErrNotFound {
			// close connection if handshake failed
			close([]byte("handshake failed, invalid key\n"))
			runtime.handshake_failed = true
			return
		} else if err != nil {
			// close connection if handshake failed
			close([]byte("internal error\n"))
			return
		}

		runtime.tenant_id = info.TenantId

		// handshake completed
		runtime.handshake = true
	} else if !runtime.registration_transferred {
		// process handle shake if not completed
		declaration, err := parser.UnmarshalJsonBytes[plugin_entities.PluginDeclaration](message)
		if err != nil {
			// close connection if handshake failed
			close([]byte("handshake failed, invalid plugin declaration\n"))
			return
		}

		runtime.Config = declaration

		// registration transferred
		runtime.registration_transferred = true
	} else if !runtime.tools_registration_transferred {
		tools, err := parser.UnmarshalJsonBytes2Slice[plugin_entities.ToolProviderDeclaration](message)
		if err != nil {
			log.Error("tools register failed, error: %v", err)
			close([]byte("tools register failed, invalid tools declaration\n"))
			return
		}

		runtime.tools_registration_transferred = true

		if len(tools) > 0 {
			declaration := runtime.Config
			declaration.Tool = &tools[0]
			runtime.Config = declaration
		}
	} else if !runtime.models_registration_transferred {
		models, err := parser.UnmarshalJsonBytes2Slice[plugin_entities.ModelProviderDeclaration](message)
		if err != nil {
			log.Error("models register failed, error: %v", err)
			close([]byte("models register failed, invalid models declaration\n"))
			return
		}

		runtime.models_registration_transferred = true

		if len(models) > 0 {
			declaration := runtime.Config
			declaration.Model = &models[0]
			runtime.Config = declaration
		}
	} else if !runtime.endpoints_registration_transferred {
		endpoints, err := parser.UnmarshalJsonBytes2Slice[plugin_entities.EndpointProviderDeclaration](message)
		if err != nil {
			log.Error("endpoints register failed, error: %v", err)
			close([]byte("endpoints register failed, invalid endpoints declaration\n"))
			return
		}

		runtime.endpoints_registration_transferred = true

		if len(endpoints) > 0 {
			declaration := runtime.Config
			declaration.Endpoint = &endpoints[0]
			runtime.Config = declaration
		}
	} else if !runtime.assets_transferred {
		assets, err := parser.UnmarshalJsonBytes2Slice[plugin_entities.RemoteAssetPayload](message)
		if err != nil {
			log.Error("assets register failed, error: %v", err)
			close([]byte("assets register failed, invalid assets declaration\n"))
			return
		}

		files := make(map[string][]byte)
		for _, asset := range assets {
			files[asset.Filename], err = hex.DecodeString(asset.Data)
			if err != nil {
				log.Error("assets decode failed, error: %v", err)
				close([]byte("assets decode failed, invalid assets data, cannot decode file\n"))
				return
			}
		}

		// remap assets
		if err := runtime.RemapAssets(&runtime.Config, files); err != nil {
			log.Error("assets remap failed, error: %v", err)
			close([]byte("assets remap failed, invalid assets data, cannot remap\n"))
			return
		}

		runtime.assets_transferred = true

		runtime.checksum = runtime.calculateChecksum()
		runtime.InitState()
		runtime.SetActiveAt(time.Now())

		// trigger registration event
		if err := runtime.Register(); err != nil {
			log.Error("register failed, error: %v", err)
			close([]byte("register failed, cannot register\n"))
			return
		}

		// publish runtime to watcher
		s.response.Write(runtime)
	} else {
		// continue handle messages if handshake completed
		runtime.response.Write(message)
	}
}
