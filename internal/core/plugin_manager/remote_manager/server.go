package remote_manager

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/media_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/panjf2000/gnet/v2"

	gnet_errors "github.com/panjf2000/gnet/v2/pkg/errors"
)

type RemotePluginServer struct {
	server *DifyServer
}

type RemotePluginServerInterface interface {
	Read() (plugin_entities.PluginFullDuplexLifetime, error)
	Next() bool
	Wrap(f func(plugin_entities.PluginFullDuplexLifetime))
	Stop() error
	Launch() error
}

// continue accepting new connections
func (r *RemotePluginServer) Read() (plugin_entities.PluginFullDuplexLifetime, error) {
	if r.server.response == nil {
		return nil, errors.New("plugin server not started")
	}

	return r.server.response.Read()
}

// Next returns true if there are more connections to be read
func (r *RemotePluginServer) Next() bool {
	if r.server.response == nil {
		return false
	}

	return r.server.response.Next()
}

// Wrap wraps the wrap method of stream response
func (r *RemotePluginServer) Wrap(f func(plugin_entities.PluginFullDuplexLifetime)) {
	r.server.response.Async(f)
}

// Stop stops the server
func (r *RemotePluginServer) Stop() error {
	if r.server.response == nil {
		return errors.New("plugin server not started")
	}
	r.server.response.Close()
	err := r.server.engine.Stop(context.Background())

	if err == gnet_errors.ErrEmptyEngine || err == gnet_errors.ErrEngineInShutdown {
		return nil
	}

	return err
}

// Launch starts the server
func (r *RemotePluginServer) Launch() error {
	// kill the process if port is already in use
	exec.Command("fuser", "-k", "tcp", fmt.Sprintf("%d", r.server.port)).Run()

	time.Sleep(time.Millisecond * 100)

	err := gnet.Run(
		r.server, r.server.addr, gnet.WithMulticore(r.server.multicore),
		gnet.WithNumEventLoop(r.server.num_loops),
	)

	if err != nil {
		r.Stop()
	}

	return err
}

// NewRemotePluginServer creates a new RemotePluginServer
func NewRemotePluginServer(config *app.Config, media_manager *media_manager.MediaBucket) *RemotePluginServer {
	addr := fmt.Sprintf(
		"tcp://%s:%d",
		config.PluginRemoteInstallingHost,
		config.PluginRemoteInstallingPort,
	)

	response := stream.NewStream[plugin_entities.PluginFullDuplexLifetime](
		config.PluginRemoteInstallingMaxConn,
	)

	multicore := true
	s := &DifyServer{
		mediaManager: media_manager,
		addr:         addr,
		port:         config.PluginRemoteInstallingPort,
		multicore:    multicore,
		num_loops:    config.PluginRemoteInstallServerEventLoopNums,
		response:     response,

		plugins:      make(map[int]*RemotePluginRuntime),
		plugins_lock: &sync.RWMutex{},

		shutdown_chan: make(chan bool),

		max_conn: int32(config.PluginRemoteInstallingMaxConn),
	}

	manager := &RemotePluginServer{
		server: s,
	}

	return manager
}
