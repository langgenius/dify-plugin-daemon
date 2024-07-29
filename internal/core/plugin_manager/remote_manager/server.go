package remote_manager

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/panjf2000/gnet/v2"

	gnet_errors "github.com/panjf2000/gnet/v2/pkg/errors"
)

type RemotePluginServer struct {
	server *DifyServer
}

// continue accepting new connections
func (r *RemotePluginServer) Read() (*RemotePluginRuntime, error) {
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
func (r *RemotePluginServer) Wrap(f func(*RemotePluginRuntime)) {
	r.server.response.Wrap(f)
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
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", r.server.port))
	if err != nil && strings.Contains(err.Error(), "address already in use") {
		scanner := bufio.NewScanner(os.Stdin)
		log.Info("Port is already in use, do you want to kill the process using the port? (y/n)")
		for scanner.Scan() {
			if scanner.Text() == "y" {
				exec.Command("fuser", "-k", "tcp", fmt.Sprintf("%d", r.server.port)).Run()
			} else if scanner.Text() == "n" {
				return errors.New("port is already in use")
			}
		}
	} else {
		listener.Close()
	}

	err = gnet.Run(
		r.server, r.server.addr, gnet.WithMulticore(r.server.multicore),
		gnet.WithNumEventLoop(r.server.num_loops),
	)

	if err != nil {
		r.Stop()
	}

	return err
}

// NewRemotePluginServer creates a new RemotePluginServer
func NewRemotePluginServer(config *app.Config) *RemotePluginServer {
	addr := fmt.Sprintf(
		"tcp://%s:%d",
		config.PluginRemoteInstallingHost,
		config.PluginRemoteInstallingPort,
	)

	response := stream.NewStreamResponse[*RemotePluginRuntime](
		config.PluginRemoteInstallingMaxConn,
	)

	multicore := true
	s := &DifyServer{
		addr:      addr,
		port:      config.PluginRemoteInstallingPort,
		multicore: multicore,
		num_loops: config.PluginRemoteInstallServerEventLoopNums,
		response:  response,

		plugins:      make(map[int]*RemotePluginRuntime),
		plugins_lock: &sync.RWMutex{},

		shutdown_chan: make(chan bool),
	}

	manager := &RemotePluginServer{
		server: s,
	}

	return manager
}
