package debugging_runtime

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	cloudoss "github.com/langgenius/dify-cloud-kit/oss"
	"github.com/langgenius/dify-cloud-kit/oss/factory"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/media_transport"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/service/debugging_service"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/constants"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/manifest_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/network"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
)

func init() {
	// init routine pool for testing
	routine.InitPool(1024)
}

func preparePluginServer(t *testing.T) (*RemotePluginServer, uint16) {
	config := &app.Config{
		DBType:     app.DB_TYPE_POSTGRESQL,
		DBUsername: "postgres",
		DBPassword: "difyai123456",
		DBHost:     "localhost",
		DBPort:     5432,
		DBDatabase: "dify_plugin_daemon",
		DBSslMode:  "disable",
	}
	config.SetDefault()
	db.Init(config)

	port, err := network.GetRandomPort()
	if err != nil {
		t.Errorf("failed to get random port: %s", err.Error())
		return nil, 0
	}
	oss, err := factory.Load("local", cloudoss.OSSArgs{
		Local: &cloudoss.Local{
			Path: "./storage",
		},
	},
	)
	if err != nil {
		t.Error("failed to load local storage", err.Error())
	}

	// start plugin server
	return NewDebuggingPluginServer(&app.Config{
		PluginRemoteInstallingHost:             "0.0.0.0",
		PluginRemoteInstallingPort:             port,
		PluginRemoteInstallingMaxConn:          1,
		PluginRemoteInstallServerEventLoopNums: 8,
	}, media_transport.NewAssetsBucket(oss, "assets", 10)), port
}

// TestLaunchAndClosePluginServer tests the launch and close of the plugin server
func TestLaunchAndClosePluginServer(t *testing.T) {
	// start plugin server
	server, _ := preparePluginServer(t)
	if server == nil {
		return
	}

	doneChan := make(chan error)

	go func() {
		err := server.Launch()
		if err != nil {
			doneChan <- err
		}
	}()

	timer := time.NewTimer(time.Second * 5)

	select {
	case err := <-doneChan:
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

type TestPluginRuntimeNotifier struct {
	onConnected func(rpr *RemotePluginRuntime) error
}

func (n *TestPluginRuntimeNotifier) OnRuntimeConnected(rpr *RemotePluginRuntime) error {
	return n.onConnected(rpr)
}

func (n *TestPluginRuntimeNotifier) OnRuntimeDisconnected(rpr *RemotePluginRuntime) {
}

func (n *TestPluginRuntimeNotifier) OnServerShutdown(reason ServerShutdownReason) {}

// TestAcceptConnection tests the acceptance of the connection
func TestAcceptConnection(t *testing.T) {
	if cache.InitRedisClient("0.0.0.0:6379", "", "difyai123456", false, 0) != nil {
		t.Errorf("failed to init redis client")
		return
	}

	tenantId := uuid.New().String()

	defer cache.Close()
	key, err := debugging_service.GetConnectionKey(debugging_service.ConnectionInfo{
		TenantId: tenantId,
	})
	if err != nil {
		t.Errorf("failed to get connection key: %s", err.Error())
		return
	}
	defer debugging_service.ClearConnectionKey(tenantId)

	server, port := preparePluginServer(t)
	if server == nil {
		return
	}
	defer server.Stop()
	go func() {
		server.Launch()
	}()

	gotConnection := false
	var connectionErr error

	server.AddNotifier(&TestPluginRuntimeNotifier{
		onConnected: func(runtime *RemotePluginRuntime) error {
			config := runtime.Configuration()
			if config.Name != "ci_test" {
				connectionErr = errors.New("plugin name not matched")
			}

			if runtime.tenantId != tenantId {
				connectionErr = errors.New("tenant id not matched")
			}

			gotConnection = true
			runtime.Stop()

			return nil
		},
	})

	// wait for the server to start
	time.Sleep(time.Second * 2)

	conn, err := net.Dial("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		t.Errorf("failed to connect to plugin server: %s", err.Error())
		return
	}

	// send handshake
	pluginManifest := parser.MarshalJsonBytes(&plugin_entities.PluginDeclaration{
		PluginDeclarationWithoutAdvancedFields: plugin_entities.PluginDeclarationWithoutAdvancedFields{
			Version: "1.0.0",
			Type:    manifest_entities.PluginType,
			Description: plugin_entities.I18nObject{
				EnUS: "test",
			},
			Author: "yeuoly",
			Name:   "ci_test",
			Icon:   "test.svg",
			Label: plugin_entities.I18nObject{
				EnUS: "ci_test",
			},
			CreatedAt: time.Now(),
			Resource: plugin_entities.PluginResourceRequirement{
				Memory:     1,
				Permission: nil,
			},
			Plugins: plugin_entities.PluginExtensions{
				Tools: []string{
					"test",
				},
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

	conn.Write(parser.MarshalJsonBytes(plugin_entities.RemotePluginRegisterPayload{
		Type: plugin_entities.REGISTER_EVENT_TYPE_HAND_SHAKE,
		Data: parser.MarshalJsonBytes(plugin_entities.RemotePluginRegisterHandshake{
			Key: key,
		}),
	})) // transfer connection key
	conn.Write([]byte("\n\n"))
	conn.Write(parser.MarshalJsonBytes(plugin_entities.RemotePluginRegisterPayload{
		Type: plugin_entities.REGISTER_EVENT_TYPE_MANIFEST_DECLARATION,
		Data: pluginManifest,
	})) // transfer manifest declaration
	conn.Write([]byte("\n\n"))
	conn.Write(parser.MarshalJsonBytes(plugin_entities.RemotePluginRegisterPayload{
		Type: plugin_entities.REGISTER_EVENT_TYPE_ENDPOINT_DECLARATION,
		Data: parser.MarshalJsonBytes([]plugin_entities.EndpointProviderDeclaration{
			{
				Settings: []plugin_entities.ProviderConfig{},
				Endpoints: []plugin_entities.EndpointDeclaration{
					{
						Path:   "/duck/<app_id>",
						Method: "GET",
					},
				},
			},
		}),
	})) // transfer endpoint declaration
	conn.Write([]byte("\n\n"))
	conn.Write(parser.MarshalJsonBytes(plugin_entities.RemotePluginRegisterPayload{
		Type: plugin_entities.REGISTER_EVENT_TYPE_ASSET_CHUNK,
		Data: parser.MarshalJsonBytes(plugin_entities.RemotePluginRegisterAssetChunk{
			Filename: "test.svg",
			Data:     "AAAA", // base64 encoded data
			End:      true,
		}),
	})) // transfer asset chunk
	conn.Write([]byte("\n\n"))
	conn.Write(parser.MarshalJsonBytes(plugin_entities.RemotePluginRegisterPayload{
		Type: plugin_entities.REGISTER_EVENT_TYPE_END,
		Data: []byte("{}"),
	})) // init process end
	conn.Write([]byte("\n\n"))
	closedChan := make(chan bool)

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
		close(closedChan)
	}()

	select {
	case <-time.After(time.Second * 10):
		// connection not closed
		t.Errorf("connection not closed normally")
		return
	case <-closedChan:
		// success

		if !gotConnection {
			t.Errorf("failed to accept connection: %s", msg)
			return
		}
		if connectionErr != nil {
			t.Errorf("failed to accept connection: %s", connectionErr.Error())
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

	server.AddNotifier(&TestPluginRuntimeNotifier{
		onConnected: func(runtime *RemotePluginRuntime) error {
			runtime.Stop()
			return nil
		},
	})

	// wait for the server to start
	time.Sleep(time.Second * 2)

	conn, err := net.Dial("tcp", fmt.Sprintf("0.0.0.0:%d", port))

	if err != nil {
		t.Errorf("failed to connect to plugin server: %s", err.Error())
		return
	}

	closedChan := make(chan bool)

	go func() {
		// block here to accept messages until the connection is closed
		buffer := make([]byte, 1024)
		for {
			_, err := conn.Read(buffer)
			if err != nil {
				break
			}
		}
		close(closedChan)
	}()

	select {
	case <-time.After(time.Second * 15):
		// connection not closed due to no handshake
		t.Errorf("connection not closed normally")
		return
	case <-closedChan:
		// success
		return
	}
}

func TestIncorrectHandshake(t *testing.T) {
	if cache.InitRedisClient("0.0.0.0:6379", "", "difyai123456", false, 0) != nil {
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

	server.AddNotifier(&TestPluginRuntimeNotifier{
		onConnected: func(runtime *RemotePluginRuntime) error {
			runtime.Stop()
			return nil
		},
	})

	// wait for the server to start
	time.Sleep(time.Second * 2)

	conn, err := net.Dial("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		t.Errorf("failed to connect to plugin server: %s", err.Error())
		return
	}

	// send incorrect handshake
	conn.Write([]byte("hello world\n"))

	closedChan := make(chan bool)
	handShakeFailed := false

	go func() {
		// block here to accept messages until the connection is closed
		buffer := make([]byte, 1024)
		for {
			_, err := conn.Read(buffer)
			if err != nil {
				break
			} else {
				if strings.Contains(string(buffer), "handshake failed") {
					handShakeFailed = true
				}
			}
		}

		close(closedChan)
	}()

	select {
	case <-time.After(time.Second * 10):
		// connection not closed
		t.Errorf("connection not closed normally")
		return
	case <-closedChan:
		if !handShakeFailed {
			t.Errorf("failed to detect incorrect handshake")
			return
		}
		return
	}
}
