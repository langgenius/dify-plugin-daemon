package debugging_runtime

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/media_transport"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/panjf2000/gnet/v2"

	gnet_errors "github.com/panjf2000/gnet/v2/pkg/errors"
)

type RemotePluginServer struct {
	server *DifyServer
}

type RemotePluginServerInterface interface {
	Stop() error
	Launch() error
}

// Stop stops the server
func (r *RemotePluginServer) Stop() error {
	err := r.server.engine.Stop(context.Background())
	if err == gnet_errors.ErrEmptyEngine || err == gnet_errors.ErrEngineInShutdown {
		return nil
	}

	return err
}

// Launch starts the server
func (r *RemotePluginServer) Launch() error {
	err := gnet.Run(
		r.server, r.server.addr, gnet.WithMulticore(r.server.multicore),
		gnet.WithNumEventLoop(r.server.numLoops),
	)

	if err != nil {
		r.Stop()
	}

	// collect shutdown signal
	go r.collectShutdownSignal()

	return err
}

func (s *RemotePluginServer) collectShutdownSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	<-c

	// shutdown server
	s.Stop()
}

// NewDebuggingPluginServer creates a new RemotePluginServer
func NewDebuggingPluginServer(
	config *app.Config, media_transport *media_transport.MediaBucket,
) *RemotePluginServer {
	addr := fmt.Sprintf(
		"tcp://%s:%d",
		config.PluginRemoteInstallingHost,
		config.PluginRemoteInstallingPort,
	)

	multicore := true
	s := &DifyServer{
		mediaManager: media_transport,
		addr:         addr,
		port:         config.PluginRemoteInstallingPort,
		multicore:    multicore,
		numLoops:     config.PluginRemoteInstallServerEventLoopNums,

		plugins:     make(map[int]*RemotePluginRuntime),
		pluginsLock: &sync.RWMutex{},

		maxConn: int32(config.PluginRemoteInstallingMaxConn),

		notifiers:     []PluginRuntimeNotifier{},
		notifierMutex: &sync.RWMutex{},
	}

	manager := &RemotePluginServer{
		server: s,
	}

	return manager
}

// AddNotifier adds a notifier to the runtime
func (r *RemotePluginServer) AddNotifier(notifier PluginRuntimeNotifier) {
	r.server.AddNotifier(notifier)
}

// WalkNotifiers walks through all the notifiers and calls the given function
func (r *RemotePluginServer) WalkNotifiers(fn func(notifier PluginRuntimeNotifier)) {
	r.server.WalkNotifiers(fn)
}
