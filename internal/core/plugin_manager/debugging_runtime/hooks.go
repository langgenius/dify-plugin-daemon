package debugging_runtime

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/basic_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/media_transport"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/panjf2000/gnet/v2"
)

var (
	// mode is only used for testing
	// TODO: simplify this ugly code
	_mode pluginRuntimeMode
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

	// read new connections
	response *stream.Stream[plugin_entities.PluginFullDuplexLifetime]

	plugins     map[int]*RemotePluginRuntime
	pluginsLock *sync.RWMutex

	shutdownChan chan bool

	maxConn     int32
	currentConn int32
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

		shutdownChan:     make(chan bool),
		waitLaunchedChan: make(chan error),

		alive: true,
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
	plugin.onDisconnected()

	// uninstall plugin
	if plugin.assetsTransferred {
		if _mode != _PLUGIN_RUNTIME_MODE_CI {
			if plugin.installationId != "" {
				if err := plugin.Unregister(); err != nil {
					log.Error("unregister plugin failed, error: %v", err)
				}
			}

			// decrease current connection
			atomic.AddInt32(&s.currentConn, -1)
		}
	}

	// send stopped event
	plugin.waitChanLock.Lock()
	for _, c := range plugin.waitStoppedChan {
		select {
		case c <- true:
		default:
		}
	}
	plugin.waitChanLock.Unlock()

	// recycle launched chan, avoid memory leak
	plugin.waitLaunchedChanOnce.Do(func() {
		close(plugin.waitLaunchedChan)
	})

	return gnet.None
}

func (s *DifyServer) OnShutdown(c gnet.Engine) {
	close(s.shutdownChan)
}

func (s *DifyServer) OnTraffic(c gnet.Conn) (action gnet.Action) {
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

		if registerPayload.Type == plugin_entities.REGISTER_EVENT_TYPE_HAND_SHAKE {
			if runtime.handshake {
				// handshake already completed
				return
			}

			key, err := parser.UnmarshalJsonBytes[plugin_entities.RemotePluginRegisterHandshake](registerPayload.Data)
			if err != nil {
				// close connection if handshake failed
				closeConn([]byte("handshake failed, invalid handshake message\n"))
				runtime.handshakeFailed = true
				return
			}

			info, err := GetConnectionInfo(key.Key)
			if err == cache.ErrNotFound {
				// close connection if handshake failed
				closeConn([]byte("handshake failed, invalid key\n"))
				runtime.handshakeFailed = true
				return
			} else if err != nil {
				// close connection if handshake failed
				log.Error("failed to get connection info: %v", err)
				closeConn([]byte("internal error\n"))
				return
			}

			runtime.tenantId = info.TenantId

			// handshake completed
			runtime.handshake = true
		} else if registerPayload.Type == plugin_entities.REGISTER_EVENT_TYPE_ASSET_CHUNK {
			if runtime.assetsTransferred {
				return
			}

			assetChunk, err := parser.UnmarshalJsonBytes[plugin_entities.RemotePluginRegisterAssetChunk](registerPayload.Data)
			if err != nil {
				log.Error("assets register failed, error: %v", err)
				closeConn([]byte("assets register failed, invalid assets chunk\n"))
				return
			}

			buffer, ok := runtime.assets[assetChunk.Filename]
			if !ok {
				runtime.assets[assetChunk.Filename] = &bytes.Buffer{}
				buffer = runtime.assets[assetChunk.Filename]
			}

			// allows at most 50MB assets
			if runtime.assetsBytes+int64(len(assetChunk.Data)) > 50*1024*1024 {
				closeConn([]byte("assets too large, at most 50MB\n"))
				return
			}

			// decode as base64
			data, err := base64.StdEncoding.DecodeString(assetChunk.Data)
			if err != nil {
				log.Error("assets decode failed, error: %v", err)
				closeConn([]byte("assets decode failed, invalid assets data\n"))
				return
			}

			buffer.Write(data)

			// update assets bytes
			runtime.assetsBytes += int64(len(data))
		} else if registerPayload.Type == plugin_entities.REGISTER_EVENT_TYPE_END {
			if !runtime.modelsRegistrationTransferred &&
				!runtime.endpointsRegistrationTransferred &&
				!runtime.toolsRegistrationTransferred &&
				!runtime.agentStrategyRegistrationTransferred {
				closeConn([]byte("no registration transferred, cannot initialize\n"))
				return
			}

			files := make(map[string][]byte)
			for filename, buffer := range runtime.assets {
				files[filename] = buffer.Bytes()
			}

			// remap assets
			if err := runtime.RemapAssets(&runtime.Config, files); err != nil {
				log.Error("assets remap failed, error: %v", err)
				closeConn([]byte(fmt.Sprintf("assets remap failed, invalid assets data, cannot remap: %v\n", err)))
				return
			}

			atomic.AddInt32(&s.currentConn, 1)
			if atomic.LoadInt32(&s.currentConn) > int32(s.maxConn) {
				closeConn([]byte("server is busy now, please try again later\n"))
				return
			}

			// fill in default values
			runtime.Config.FillInDefaultValues()

			// mark assets transferred
			runtime.assetsTransferred = true

			runtime.checksum = runtime.calculateChecksum()
			runtime.InitState()
			runtime.SetActiveAt(time.Now())

			// trigger registration event
			if err := runtime.Register(); err != nil {
				closeConn([]byte(fmt.Sprintf("register failed, cannot register: %v\n", err)))
				return
			}

			// send started event
			runtime.waitChanLock.Lock()
			for _, c := range runtime.waitStartedChan {
				select {
				case c <- true:
				default:
				}
			}
			runtime.waitChanLock.Unlock()

			// notify launched
			runtime.waitLaunchedChanOnce.Do(func() {
				close(runtime.waitLaunchedChan)
			})

			// mark initialized
			runtime.initialized = true

			// publish runtime to watcher
			s.response.Write(runtime)
		} else if registerPayload.Type == plugin_entities.REGISTER_EVENT_TYPE_MANIFEST_DECLARATION {
			if runtime.registrationTransferred {
				return
			}

			// process handle shake if not completed
			declaration, err := parser.UnmarshalJsonBytes[plugin_entities.PluginDeclaration](registerPayload.Data)
			if err != nil {
				// close connection if handshake failed
				closeConn([]byte(fmt.Sprintf("handshake failed, invalid plugin declaration: %v\n", err)))
				return
			}

			runtime.Config = declaration

			// registration transferred
			runtime.registrationTransferred = true
		} else if registerPayload.Type == plugin_entities.REGISTER_EVENT_TYPE_TOOL_DECLARATION {
			if runtime.toolsRegistrationTransferred {
				return
			}

			tools, err := parser.UnmarshalJsonBytes2Slice[plugin_entities.ToolProviderDeclaration](registerPayload.Data)
			if err != nil {
				closeConn([]byte(fmt.Sprintf("tools register failed, invalid tools declaration: %v\n", err)))
				return
			}

			runtime.toolsRegistrationTransferred = true

			if len(tools) > 0 {
				declaration := runtime.Config
				declaration.Tool = &tools[0]
				runtime.Config = declaration
			}
		} else if registerPayload.Type == plugin_entities.REGISTER_EVENT_TYPE_MODEL_DECLARATION {
			if runtime.modelsRegistrationTransferred {
				return
			}

			models, err := parser.UnmarshalJsonBytes2Slice[plugin_entities.ModelProviderDeclaration](registerPayload.Data)
			if err != nil {
				closeConn([]byte(fmt.Sprintf("models register failed, invalid models declaration: %v\n", err)))
				return
			}

			runtime.modelsRegistrationTransferred = true

			if len(models) > 0 {
				declaration := runtime.Config
				declaration.Model = &models[0]
				runtime.Config = declaration
			}
		} else if registerPayload.Type == plugin_entities.REGISTER_EVENT_TYPE_ENDPOINT_DECLARATION {
			if runtime.endpointsRegistrationTransferred {
				return
			}

			endpoints, err := parser.UnmarshalJsonBytes2Slice[plugin_entities.EndpointProviderDeclaration](registerPayload.Data)
			if err != nil {
				closeConn([]byte(fmt.Sprintf("endpoints register failed, invalid endpoints declaration: %v\n", err)))
				return
			}

			runtime.endpointsRegistrationTransferred = true

			if len(endpoints) > 0 {
				declaration := runtime.Config
				declaration.Endpoint = &endpoints[0]
				runtime.Config = declaration
			}
		} else if registerPayload.Type == plugin_entities.REGISTER_EVENT_TYPE_AGENT_STRATEGY_DECLARATION {
			if runtime.agentStrategyRegistrationTransferred {
				return
			}

			agents, err := parser.UnmarshalJsonBytes2Slice[plugin_entities.AgentStrategyProviderDeclaration](registerPayload.Data)
			if err != nil {
				closeConn([]byte(fmt.Sprintf("agent strategies register failed, invalid agent strategies declaration: %v\n", err)))
				return
			}

			runtime.agentStrategyRegistrationTransferred = true

			if len(agents) > 0 {
				declaration := runtime.Config
				declaration.AgentStrategy = &agents[0]
				runtime.Config = declaration
			}
		}
	} else {
		// continue handle messages if handshake completed
		runtime.response.Write(message)
	}
}
