package debugging_runtime

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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

// Stop stops the server gracefully
func (r *RemotePluginServer) Stop() error {
	if r.server == nil {
		return nil
	}

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := r.server.engine.Stop(ctx)
	if err == gnet_errors.ErrEmptyEngine || err == gnet_errors.ErrEngineInShutdown {
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to stop server gracefully: %w", err)
	}

	return nil
}

// Launch starts the server
func (r *RemotePluginServer) Launch() error {
	// Try to start the server with retry mechanism
	// This handles the case where the port is in TIME_WAIT state after a crash
	maxRetries := 3
	var err error

	for i := 0; i < maxRetries; i++ {
		err = gnet.Run(
			r.server, r.server.addr,
			gnet.WithMulticore(r.server.multicore),
			gnet.WithNumEventLoop(r.server.numLoops),
			gnet.WithLogger(GnetLogger{}),
			gnet.WithReuseAddr(true),
			gnet.WithReusePort(true),
		)

		if err == nil {
			break
		}

		// If this is the last retry, don't wait
		if i < maxRetries-1 {
			waitTime := (i + 1) * 2
			GnetLogger{}.Warnf("Failed to bind to %s (attempt %d/%d): %v, retrying in %d seconds...\n",
				r.server.addr, i+1, maxRetries, err, waitTime)
			time.Sleep(time.Duration(waitTime) * time.Second)
		}
	}

	if err != nil {
		err := r.Stop()
		if err != nil {
			return err
		}
		return fmt.Errorf("failed to start server after %d attempts: %w", maxRetries, err)
	}

	// collect shutdown signal
	go r.collectShutdownSignal()

	return nil
}

func (r *RemotePluginServer) collectShutdownSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	sig := <-c
	fmt.Printf("\nReceived signal %v, shutting down server gracefully...\n", sig)

	// shutdown server with timeout
	if err := r.Stop(); err != nil {
		fmt.Printf("Error shutting down server: %v\n", err)
	} else {
		fmt.Println("Server shut down successfully")
	}
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
