package remote_manager

import (
	"sync"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/panjf2000/gnet/v2"
)

type DifyServer struct {
	gnet.BuiltinEventEngine

	engine gnet.Engine

	// listening address
	addr string

	// enabled multicore
	multicore bool

	// event loop count
	num_loops int

	// read new connections
	response *stream.StreamResponse[*RemotePluginRuntime]

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

	s.response.Write(runtime)

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
	plugin.close()

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
		runtime.response.Write(message)
	}

	return gnet.None
}
