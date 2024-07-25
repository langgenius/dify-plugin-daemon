package remote_manager

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
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
	var connection_err error

	go func() {
		for server.Next() {
			runtime, err := server.Read()
			if err != nil {
				t.Errorf("failed to read plugin runtime: %s", err.Error())
				return
			}

			if runtime.Config.Name != "ci_test" {
				connection_err = errors.New("plugin name not matched")
			}

			got_connection = true
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

	// send handshake
	handle_shake_message := parser.MarshalJsonBytes(&plugin_entities.PluginDeclaration{
		Version:   "1.0.0",
		Type:      plugin_entities.PluginType,
		Author:    "Yeuoly",
		Name:      "ci_test",
		CreatedAt: time.Now(),
		Resource: plugin_entities.PluginResourceRequirement{
			Memory:     1,
			Storage:    1,
			Permission: nil,
		},
		Plugins: []string{
			"test",
		},
		Execution: plugin_entities.PluginDeclarationExecution{
			Install: "echo 'hello'",
			Launch:  "echo 'hello'",
		},
	})
	conn.Write(handle_shake_message)
	conn.Write([]byte("\n"))
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
		if connection_err != nil {
			t.Errorf("failed to accept connection: %s", connection_err.Error())
			return
		}
		return
	}
}

func TestNoHandleShakeIn10Seconds(t *testing.T) {
	server, port := preparePluginServer(t)
	if server == nil {
		return
	}
	defer server.Stop()
	go func() {
		server.Launch()
	}()

	go func() {
		for server.Next() {
			runtime, err := server.Read()
			if err != nil {
				t.Errorf("failed to read plugin runtime: %s", err.Error())
				return
			}

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
	case <-time.After(time.Second * 15):
		// connection not closed due to no handshake
		t.Errorf("connection not closed normally")
		return
	case <-closed_chan:
		// success
		return
	}
}

func TestIncorrectHandshake(t *testing.T) {
	server, port := preparePluginServer(t)
	if server == nil {
		return
	}
	defer server.Stop()
	go func() {
		server.Launch()
	}()

	go func() {
		for server.Next() {
			runtime, err := server.Read()
			if err != nil {
				t.Errorf("failed to read plugin runtime: %s", err.Error())
				return
			}

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

	// send incorrect handshake
	conn.Write([]byte("hello world\n"))

	closed_chan := make(chan bool)
	hand_shake_failed := false

	go func() {
		// block here to accept messages until the connection is closed
		buffer := make([]byte, 1024)
		for {
			_, err := conn.Read(buffer)
			if err != nil {
				break
			} else {
				if strings.Contains(string(buffer), "handshake failed") {
					hand_shake_failed = true
				}
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
		if !hand_shake_failed {
			t.Errorf("failed to detect incorrect handshake")
			return
		}
		return
	}
}
