package remote_manager

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
)

func preparePluginServer(t *testing.T) (*RemotePluginServer, uint16) {
	// generate a random port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Errorf("failed to get a random port: %s", err.Error())
		return nil, 0
	}
	listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port

	// start plugin server
	return NewRemotePluginServer(&app.Config{
		PluginRemoteInstallingHost:             "0.0.0.0",
		PluginRemoteInstallingPort:             uint16(port),
		PluginRemoteInstallingMaxConn:          1,
		PluginRemoteInstallServerEventLoopNums: 8,
	}), uint16(port)
}

// TestLaunchAndClosePluginServer tests the launch and close of the plugin server
func TestLaunchAndClosePluginServer(t *testing.T) {
	// start plugin server
	server, _ := preparePluginServer(t)
	if server == nil {
		return
	}

	done_chan := make(chan error)

	go func() {
		err := server.Launch()
		if err != nil {
			done_chan <- err
		}
	}()

	timer := time.NewTimer(time.Second * 5)

	select {
	case err := <-done_chan:
		t.Errorf("failed to launch plugin server: %s", err.Error())
		return
	case <-timer.C:
		err := server.Stop()
		if err != nil {
			t.Errorf("failed to stop plugin server: %s", err.Error())
			return
		}
	}
}

// TestAcceptConnection tests the acceptance of the connection
func TestAcceptConnection(t *testing.T) {
	server, port := preparePluginServer(t)
	if server == nil {
		return
	}
	defer server.Stop()
	go func() {
		server.Launch()
	}()

	got_connection := false

	go func() {
		for server.Next() {
			runtime, err := server.Read()
			if err != nil {
				t.Errorf("failed to read plugin runtime: %s", err.Error())
				return
			}

			got_connection = true

			time.Sleep(time.Second * 2)
			runtime.Stop()
		}
	}()

	// wait for the server to start
	time.Sleep(time.Second * 2)

	conn, err := net.Dial("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		t.Errorf("failed to connect to plugin server: %s", err.Error())
		return
	}

	closed_chan := make(chan bool)

	go func() {
		// block here to accept messages until the connection is closed
		buffer := make([]byte, 1024)
		for {
			_, err := conn.Read(buffer)
			if err != nil {
				break
			}
		}
		close(closed_chan)
	}()

	select {
	case <-time.After(time.Second * 10):
		// connection not closed
		t.Errorf("connection not closed normally")
		return
	case <-closed_chan:
		// success
		if !got_connection {
			t.Errorf("failed to accept connection")
			return
		}
		return
	}
}
