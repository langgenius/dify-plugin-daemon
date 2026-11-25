package service

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"sync"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	cloudoss "github.com/langgenius/dify-cloud-kit/oss"
	"github.com/langgenius/dify-cloud-kit/oss/factory"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	serverless "github.com/langgenius/dify-plugin-daemon/internal/core/serverless_connector"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUpgradePlugin(t *testing.T) {
	originalIdentifier, err := plugin_entities.NewPluginUniqueIdentifier("author/test-plugin:1.0.0@abcdef1234567890abcdef1234567890ab")
	if err != nil {
		t.Fatalf("failed to create original plugin unique identifier: %v", err)
	}

	newIdentifier, err := plugin_entities.NewPluginUniqueIdentifier("author/test-plugin:2.0.0@1234567890abcdef1234567890abcdef12")
	if err != nil {
		t.Fatalf("failed to create new plugin unique identifier: %v", err)
	}

	config := &app.Config{
		PluginInstallTimeout: 15,
	}

	tests := []struct {
		name                             string
		tenantId                         string
		source                           string
		meta                             map[string]any
		originalPluginUniqueIdentifier   plugin_entities.PluginUniqueIdentifier
		newPluginUniqueIdentifier        plugin_entities.PluginUniqueIdentifier
		wantSuccess                      bool
		wantAllInstalled                 bool
		wantTaskIDEmpty                  bool
	}{
		{
			name:                           "same plugin identifiers",
			tenantId:                       "tenant-123",
			source:                         "test",
			meta:                           map[string]any{},
			originalPluginUniqueIdentifier:  originalIdentifier,
			newPluginUniqueIdentifier:       originalIdentifier,
			wantSuccess:                    false,
		},
		{
			name:                           "different plugin identifiers",
			tenantId:                       "tenant-123",
			source:                         "test",
			meta:                           map[string]any{},
			originalPluginUniqueIdentifier:  originalIdentifier,
			newPluginUniqueIdentifier:       newIdentifier,
			wantSuccess:                    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := UpgradePlugin(
				config,
				tt.tenantId,
				tt.source,
				tt.meta,
				tt.originalPluginUniqueIdentifier,
				tt.newPluginUniqueIdentifier,
			)

			isSuccess := response.Code == 0
			if isSuccess != tt.wantSuccess {
				t.Errorf("UpgradePlugin() success = %v, want %v", isSuccess, tt.wantSuccess)
			}

			if isSuccess {
				var result InstallPluginResponse
				dataBytes, err := json.Marshal(response.Data)
				if err == nil {
					_ = json.Unmarshal(dataBytes, &result)
					if tt.wantAllInstalled && !result.AllInstalled {
						t.Errorf("UpgradePlugin() AllInstalled = %v, want %v", result.AllInstalled, tt.wantAllInstalled)
					}
					if tt.wantTaskIDEmpty && result.TaskID != "" {
						t.Errorf("UpgradePlugin() TaskID = %v, want empty", result.TaskID)
					}
				}
			}
		})
	}
}

func TestReinstallPluginSwitchesRuntimeAfterCallerDisconnects(t *testing.T) {
	routine.InitPool(16)

	redisServer := miniredis.RunT(t)
	require.NoError(t, cache.InitRedisClient(redisServer.Addr(), cache.RedisCredentials{}, false, 0, nil))
	t.Cleanup(func() {
		_ = cache.Close()
	})

	dsn := "file:" + strings.ReplaceAll(t.Name(), "/", "_") + "?mode=memory&cache=shared&_busy_timeout=5000"
	gormDB, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, gormDB.AutoMigrate(&models.ServerlessRuntime{}))
	sqlDB, err := gormDB.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	db.DifyPluginDB = gormDB
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	packageBytes, err := os.ReadFile("../core/plugin_manager/testdata/openai.difypkg")
	require.NoError(t, err)
	packageDecoder, err := decoder.NewZipPluginDecoder(packageBytes)
	require.NoError(t, err)
	identifier, err := packageDecoder.UniqueIdentity()
	require.NoError(t, err)

	require.NoError(t, db.Create(&models.ServerlessRuntime{
		PluginUniqueIdentifier: identifier.String(),
		Type:                   models.SERVERLESS_RUNTIME_TYPE_SERVERLESS,
		FunctionName:           "plugin",
		FunctionURL:            "http://svc-plugin",
	}))

	release := make(chan struct{})
	connector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ping" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`"pong"`))
			return
		}
		if r.URL.Path != "/v1/launch" {
			t.Errorf("unexpected connector request: %s", r.URL)
			return
		}
		writeChunk := func(stage, message string) {
			fmt.Fprintf(w, "data: {\"Stage\":%q,\"State\":\"success\",\"Message\":%q}\n\n", stage, message)
			w.(http.Flusher).Flush()
		}
		writeChunk("start", "")
		select {
		case <-release:
		case <-r.Context().Done():
			return
		}
		writeChunk("run", "endpoint=http://svc-plugin--2,name=plugin--2,id=plugin--2")
		writeChunk("end", "")
	}))
	t.Cleanup(connector.Close)
	releaseConnector := sync.OnceFunc(func() { close(release) })
	t.Cleanup(releaseConnector)

	key := "test-key"
	config := &app.Config{
		Platform:                                   app.PLATFORM_SERVERLESS,
		DifyPluginServerlessConnectorURL:           &connector.URL,
		DifyPluginServerlessConnectorAPIKey:        &key,
		DifyPluginServerlessConnectorLaunchTimeout: 5,
		PluginMediaCachePath:                       "assets",
		PluginMediaCacheSize:                       4,
		PluginAssetCacheSize:                       4,
		PluginInstalledPath:                        "installed",
		PluginPackageCachePath:                     "packages",
	}
	serverless.Init(config)

	storage, err := factory.Load("local", cloudoss.OSSArgs{
		Local: &cloudoss.Local{Path: t.TempDir()},
	})
	require.NoError(t, err)
	require.NoError(t, storage.Save(path.Join(config.PluginPackageCachePath, identifier.String()), packageBytes))
	plugin_manager.InitGlobalManager(storage, config)

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	handlerReturned := make(chan struct{})
	engine.POST("/reinstall", func(c *gin.Context) {
		defer close(handlerReturned)
		ReinstallPluginFromIdentifier(c, config, identifier)
	})
	daemon := httptest.NewServer(engine)
	t.Cleanup(daemon.Close)

	reqCtx, cancel := context.WithCancel(context.Background())
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, daemon.URL+"/reinstall", nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	firstLine, err := bufio.NewReader(resp.Body).ReadString('\n')
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(firstLine, "data:"), firstLine)
	cancel()
	_ = resp.Body.Close()

	select {
	case <-handlerReturned:
	case <-time.After(5 * time.Second):
		t.Fatal("reinstall handler did not observe the caller disconnect")
	}
	releaseConnector()

	require.Eventually(t, func() bool {
		runtime, err := db.GetOne[models.ServerlessRuntime](
			db.Equal("plugin_unique_identifier", identifier.String()),
		)
		return err == nil && runtime.FunctionName == "plugin--2" && runtime.FunctionURL == "http://svc-plugin--2"
	}, 5*time.Second, 20*time.Millisecond)
}
