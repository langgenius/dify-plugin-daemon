package debugging_runtime

import (
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	cloudoss "github.com/langgenius/dify-cloud-kit/oss"
	"github.com/langgenius/dify-cloud-kit/oss/factory"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/media_transport"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/constants"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/manifest_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/network"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
)

// Registration messages sent before a completed handshake must be rejected
// and the connection must fail closed without registering any runtime.
func TestRegistrationRejectedBeforeHandshake(t *testing.T) {
	port, err := network.GetRandomPort()
	if err != nil {
		t.Fatalf("failed to get random port: %s", err.Error())
	}

	oss, err := factory.Load("local", cloudoss.OSSArgs{
		Local: &cloudoss.Local{
			Path: t.TempDir(),
		},
	})
	if err != nil {
		t.Fatalf("failed to load local storage: %s", err.Error())
	}

	server := NewDebuggingPluginServer(&app.Config{
		PluginRemoteInstallingHost:             "127.0.0.1",
		PluginRemoteInstallingPort:             port,
		PluginRemoteInstallingMaxConn:          10,
		PluginRemoteInstallServerEventLoopNums: 1,
	}, media_transport.NewAssetsBucket(oss, "assets", 10))
	defer server.Stop()
	go server.Launch()

	registered := make(chan struct{}, 1)
	server.AddNotifier(&TestPluginRuntimeNotifier{
		onConnected: func(runtime *RemotePluginRuntime) error {
			registered <- struct{}{}
			return nil
		},
	})

	// wait for the server to start
	time.Sleep(time.Second * 2)

	manifest := parser.MarshalJsonBytes(&plugin_entities.PluginDeclaration{
		PluginDeclarationWithoutAdvancedFields: plugin_entities.PluginDeclarationWithoutAdvancedFields{
			Version: "1.0.0",
			Type:    manifest_entities.PluginType,
			Description: plugin_entities.I18nObject{
				EnUS: "test",
			},
			Author:    "test",
			Name:      "pre_handshake_test",
			Icon:      "icon.svg",
			Label:     plugin_entities.I18nObject{EnUS: "test"},
			CreatedAt: time.Now(),
			Resource:  plugin_entities.PluginResourceRequirement{Memory: 1},
			Plugins: plugin_entities.PluginExtensions{
				Endpoints: []string{"test"},
			},
			Meta: plugin_entities.PluginMeta{
				Version: "0.0.1",
				Arch:    []constants.Arch{constants.AMD64},
				Runner: plugin_entities.PluginRunner{
					Language:   constants.Python,
					Version:    "3.12",
					Entrypoint: "main",
				},
			},
		},
	})

	messages := []struct {
		name    string
		payload plugin_entities.RemotePluginRegisterPayload
	}{
		{
			name: "manifest declaration",
			payload: plugin_entities.RemotePluginRegisterPayload{
				Type: plugin_entities.REGISTER_EVENT_TYPE_MANIFEST_DECLARATION,
				Data: manifest,
			},
		},
		{
			name: "endpoint declaration",
			payload: plugin_entities.RemotePluginRegisterPayload{
				Type: plugin_entities.REGISTER_EVENT_TYPE_ENDPOINT_DECLARATION,
				Data: parser.MarshalJsonBytes([]plugin_entities.EndpointProviderDeclaration{
					{
						Settings: []plugin_entities.ProviderConfig{},
						Endpoints: []plugin_entities.EndpointDeclaration{
							{Path: "/test", Method: "GET"},
						},
					},
				}),
			},
		},
		{
			name: "asset chunk",
			payload: plugin_entities.RemotePluginRegisterPayload{
				Type: plugin_entities.REGISTER_EVENT_TYPE_ASSET_CHUNK,
				Data: parser.MarshalJsonBytes(plugin_entities.RemotePluginRegisterAssetChunk{
					Filename: "icon.svg",
					Data:     "QUJD",
					End:      true,
				}),
			},
		},
		{
			name: "initialization end",
			payload: plugin_entities.RemotePluginRegisterPayload{
				Type: plugin_entities.REGISTER_EVENT_TYPE_END,
				Data: []byte("{}"),
			},
		},
	}

	for _, tt := range messages {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
			if err != nil {
				t.Fatalf("failed to connect to plugin server: %s", err.Error())
			}
			defer conn.Close()

			// NOTE: no handshake message, straight to registration
			if _, err := conn.Write(parser.MarshalJsonBytes(tt.payload)); err != nil {
				t.Fatalf("failed to write payload: %s", err.Error())
			}
			if _, err := conn.Write([]byte("\n\n")); err != nil {
				t.Fatalf("failed to write delimiter: %s", err.Error())
			}

			// the server must reject the message and close the connection
			conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			var reply strings.Builder
			buf := make([]byte, 1024)
			for {
				n, err := conn.Read(buf)
				if n > 0 {
					reply.Write(buf[:n])
				}
				if err != nil {
					break
				}
			}
			if !strings.Contains(reply.String(), "handshake failed") {
				t.Errorf("expected handshake failure reply, got %q", reply.String())
			}
		})
	}

	// no runtime may be registered
	time.Sleep(time.Second)
	select {
	case <-registered:
		t.Fatal("runtime registered without a completed handshake")
	default:
	}
}
