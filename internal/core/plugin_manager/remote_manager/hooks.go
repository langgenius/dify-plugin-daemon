package remote_manager

import (
	"bytes"
	"encoding/base64"
	"fmt"
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
	// mode is only used for testing
	_mode pluginRuntimeMode
)

type DifyServer struct {
	gnet.BuiltinEventEngine

	engine gnet.Engine

	mediaManager *media_manager.MediaBucket

	// listening address
	addr string
	port uint16

	// enabled multicore
	multicore bool

	// event loop count
	num_loops int

	// read new connections
	response *stream.Stream[plugin_entities.PluginFullDuplexLifetime]

	plugins      map[int]*RemotePluginRuntime
	plugins_lock *sync.RWMutex

	shutdown_chan chan bool

	max_conn     int32
	current_conn int32
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
		response:       stream.NewStream[[]byte](512),
		callbacks:      make(map[string][]func([]byte)),
		callbacks_lock: &sync.RWMutex{},

		assets:       make(map[string]*bytes.Buffer),
		assets_bytes: 0,

		shutdown_chan:      make(chan bool),
		wait_launched_chan: make(chan error),

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

	if plugin == nil {
		return gnet.None
	}

	// close plugin
	plugin.onDisconnected()

	// uninstall plugin
	if plugin.assets_transferred {
		if _mode != _PLUGIN_RUNTIME_MODE_CI {
			if err := plugin.Unregister(); err != nil {
				log.Error("unregister plugin failed, error: %v", err)
			}

			// decrease current connection
			atomic.AddInt32(&s.current_conn, -1)
		}
	}

	// send stopped event
	plugin.wait_chan_lock.Lock()
	for _, c := range plugin.wait_stopped_chan {
		select {
		case c <- true:
		default:
		}
	}
	plugin.wait_chan_lock.Unlock()

	// recycle launched chan, avoid memory leak
	plugin.wait_launched_chan_once.Do(func() {
		close(plugin.wait_launched_chan)
	})

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

	close_conn := func(message []byte) {
		if atomic.CompareAndSwapInt32(&runtime.closed, 0, 1) {
			runtime.conn.Write(message)
			runtime.conn.Close()
		}
	}

	if !runtime.initialized {
		register_payload, err := parser.UnmarshalJsonBytes[plugin_entities.RemotePluginRegisterPayload](message)
		if err != nil {
			// close connection if handshake failed
			close_conn([]byte("handshake failed, invalid handshake message\n"))
			runtime.handshake_failed = true
			return
		}

		if register_payload.Type == plugin_entities.REGISTER_EVENT_TYPE_HAND_SHAKE {
			if runtime.handshake {
				// handshake already completed
				return
			}

			key, err := parser.UnmarshalJsonBytes[plugin_entities.RemotePluginRegisterHandshake](register_payload.Data)
			if err != nil {
				// close connection if handshake failed
				close_conn([]byte("handshake failed, invalid handshake message\n"))
				runtime.handshake_failed = true
				return
			}

			info, err := GetConnectionInfo(key.Key)
			if err == cache.ErrNotFound {
				// close connection if handshake failed
				close_conn([]byte("handshake failed, invalid key\n"))
				runtime.handshake_failed = true
				return
			} else if err != nil {
				// close connection if handshake failed
				close_conn([]byte("internal error\n"))
				return
			}

			runtime.tenant_id = info.TenantId

			// handshake completed
			runtime.handshake = true
		} else if register_payload.Type == plugin_entities.REGISTER_EVENT_TYPE_ASSET_CHUNK {
			if runtime.assets_transferred {
				return
			}

			asset_chunk, err := parser.UnmarshalJsonBytes[plugin_entities.RemotePluginRegisterAssetChunk](register_payload.Data)
			if err != nil {
				log.Error("assets register failed, error: %v", err)
				close_conn([]byte("assets register failed, invalid assets chunk\n"))
				return
			}

			buffer, ok := runtime.assets[asset_chunk.Filename]
			if !ok {
				runtime.assets[asset_chunk.Filename] = &bytes.Buffer{}
				buffer = runtime.assets[asset_chunk.Filename]
			}

			// allows at most 50MB assets
			if runtime.assets_bytes+int64(len(asset_chunk.Data)) > 50*1024*1024 {
				close_conn([]byte("assets too large, at most 50MB\n"))
				return
			}

			// decode as base64
			data, err := base64.StdEncoding.DecodeString(asset_chunk.Data)
			if err != nil {
				log.Error("assets decode failed, error: %v", err)
				close_conn([]byte("assets decode failed, invalid assets data\n"))
				return
			}

			buffer.Write(data)

			// update assets bytes
			runtime.assets_bytes += int64(len(data))
		} else if register_payload.Type == plugin_entities.REGISTER_EVENT_TYPE_END {
			if !runtime.models_registration_transferred &&
				!runtime.endpoints_registration_transferred &&
				!runtime.tools_registration_transferred {
				close_conn([]byte("no registration transferred, cannot initialize\n"))
				return
			}

			files := make(map[string][]byte)
			for filename, buffer := range runtime.assets {
				files[filename] = buffer.Bytes()
			}

			// remap assets
			if err := runtime.RemapAssets(&runtime.Config, files); err != nil {
				log.Error("assets remap failed, error: %v", err)
				close_conn([]byte(fmt.Sprintf("assets remap failed, invalid assets data, cannot remap: %v\n", err)))
				return
			}

			atomic.AddInt32(&s.current_conn, 1)
			if atomic.LoadInt32(&s.current_conn) > int32(s.max_conn) {
				close_conn([]byte("server is busy now, please try again later\n"))
				return
			}

			// fill in default values
			runtime.Config.FillInDefaultValues()

			// mark assets transferred
			runtime.assets_transferred = true

			runtime.checksum = runtime.calculateChecksum()
			runtime.InitState()
			runtime.SetActiveAt(time.Now())

			// trigger registration event
			if err := runtime.Register(); err != nil {
				log.Error("register failed, error: %v", err)
				close_conn([]byte("register failed, cannot register\n"))
				return
			}

			// send started event
			runtime.wait_chan_lock.Lock()
			for _, c := range runtime.wait_started_chan {
				select {
				case c <- true:
				default:
				}
			}
			runtime.wait_chan_lock.Unlock()

			// notify launched
			runtime.wait_launched_chan_once.Do(func() {
				close(runtime.wait_launched_chan)
			})

			// mark initialized
			runtime.initialized = true

			// publish runtime to watcher
			s.response.Write(runtime)
		} else if register_payload.Type == plugin_entities.REGISTER_EVENT_TYPE_MANIFEST_DECLARATION {
			if runtime.registration_transferred {
				return
			}

			// process handle shake if not completed
			declaration, err := parser.UnmarshalJsonBytes[plugin_entities.PluginDeclaration](register_payload.Data)
			if err != nil {
				// close connection if handshake failed
				close_conn([]byte("handshake failed, invalid plugin declaration\n"))
				return
			}

			runtime.Config = declaration

			// registration transferred
			runtime.registration_transferred = true
		} else if register_payload.Type == plugin_entities.REGISTER_EVENT_TYPE_TOOL_DECLARATION {
			if runtime.tools_registration_transferred {
				return
			}

			tools, err := parser.UnmarshalJsonBytes2Slice[plugin_entities.ToolProviderDeclaration](register_payload.Data)
			if err != nil {
				log.Error("tools register failed, error: %v", err)
				close_conn([]byte("tools register failed, invalid tools declaration\n"))
				return
			}

			runtime.tools_registration_transferred = true

			if len(tools) > 0 {
				declaration := runtime.Config
				declaration.Tool = &tools[0]
				runtime.Config = declaration
			}
		} else if register_payload.Type == plugin_entities.REGISTER_EVENT_TYPE_MODEL_DECLARATION {
			if runtime.models_registration_transferred {
				return
			}

			models, err := parser.UnmarshalJsonBytes2Slice[plugin_entities.ModelProviderDeclaration](register_payload.Data)
			if err != nil {
				log.Error("models register failed, error: %v", err)
				close_conn([]byte("models register failed, invalid models declaration\n"))
				return
			}

			runtime.models_registration_transferred = true

			if len(models) > 0 {
				declaration := runtime.Config
				declaration.Model = &models[0]
				runtime.Config = declaration
			}
		} else if register_payload.Type == plugin_entities.REGISTER_EVENT_TYPE_ENDPOINT_DECLARATION {
			if runtime.endpoints_registration_transferred {
				return
			}

			endpoints, err := parser.UnmarshalJsonBytes2Slice[plugin_entities.EndpointProviderDeclaration](register_payload.Data)
			if err != nil {
				log.Error("endpoints register failed, error: %v", err)
				close_conn([]byte("endpoints register failed, invalid endpoints declaration\n"))
				return
			}

			runtime.endpoints_registration_transferred = true

			if len(endpoints) > 0 {
				declaration := runtime.Config
				declaration.Endpoint = &endpoints[0]
				runtime.Config = declaration
			}
		} else {
			// unknown event type
			close_conn([]byte("unknown initialization event type\n"))
			return
		}
	} else {
		// continue handle messages if handshake completed
		runtime.response.Write(message)
	}
}
