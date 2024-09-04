package remote_manager

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/constants"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/network"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

func preparePluginServer(t *testing.T) (*RemotePluginServer, uint16) {
	db.Init(&app.Config{
		DBUsername: "postgres",
		DBPassword: "difyai123456",
		DBHost:     "localhost",
		DBPort:     5432,
		DBDatabase: "dify_plugin_daemon",
		DBSslMode:  "disable",
	})

	port, err := network.GetRandomPort()
	if err != nil {
		t.Errorf("failed to get random port: %s", err.Error())
		return nil, 0
	}

	// start plugin server
	return NewRemotePluginServer(&app.Config{
		PluginRemoteInstallingHost:             "0.0.0.0",
		PluginRemoteInstallingPort:             port,
		PluginRemoteInstallingMaxConn:          1,
		PluginRemoteInstallServerEventLoopNums: 8,
	}, nil), port
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
	if cache.InitRedisClient("0.0.0.0:6379", "difyai123456") != nil {
		t.Errorf("failed to init redis client")
		return
	}

	defer cache.Close()
	key, err := GetConnectionKey(ConnectionInfo{
		TenantId: "test",
	})
	if err != nil {
		t.Errorf("failed to get connection key: %s", err.Error())
		return
	}
	defer ClearConnectionKey("test")

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

			if runtime.tenant_id != "test" {
				connection_err = errors.New("tenant id not matched")
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
		PluginDeclarationWithoutAdvancedFields: plugin_entities.PluginDeclarationWithoutAdvancedFields{
			Version: "1.0.0",
			Type:    plugin_entities.PluginType,
			Author:  "Yeuoly",
			Name:    "ci_test",
			Label: plugin_entities.I18nObject{
				EnUS: "ci_test",
			},
			CreatedAt: time.Now(),
			Resource: plugin_entities.PluginResourceRequirement{
				Memory:     1,
				Permission: nil,
			},
			Plugins: []string{
				"test",
			},
			Meta: plugin_entities.PluginMeta{
				Version: "0.0.1",
				Arch: []constants.Arch{
					constants.AMD64,
				},
				Runner: plugin_entities.PluginRunner{
					Language:   constants.Python,
					Version:    "3.12",
					Entrypoint: "main",
				},
			},
		},
	})
	conn.Write([]byte(key))
	conn.Write([]byte("\n"))
	conn.Write(handle_shake_message)
	conn.Write([]byte("\n"))
	conn.Write([]byte("[]\n"))
	conn.Write([]byte("[]\n"))
	conn.Write([]byte("[]\n"))
	closed_chan := make(chan bool)

	msg := ""

	go func() {
		// block here to accept messages until the connection is closed
		buffer := make([]byte, 1024)
		for {
			n, err := conn.Read(buffer)
			if err != nil {
				break
			}
			msg += string(buffer[:n])
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
			t.Errorf("failed to accept connection: %s", msg)
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
	if cache.InitRedisClient("0.0.0.0:6379", "difyai123456") != nil {
		t.Errorf("failed to init redis client")
		return
	}

	defer cache.Close()

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
