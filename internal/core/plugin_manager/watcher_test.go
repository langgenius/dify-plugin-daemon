package plugin_manager

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/positive_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/oss/local"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

type fakePlugin struct {
	plugin_entities.PluginRuntime
	positive_manager.PositivePluginRuntime
}

func (r *fakePlugin) InitEnvironment() error {
	return nil
}

func (r *fakePlugin) Checksum() (string, error) {
	return "", nil
}

func (r *fakePlugin) Identity() (plugin_entities.PluginUniqueIdentifier, error) {
	return plugin_entities.PluginUniqueIdentifier(""), nil
}

func (r *fakePlugin) StartPlugin() error {
	return nil
}

func (r *fakePlugin) Type() plugin_entities.PluginRuntimeType {
	return plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL
}

func (r *fakePlugin) Wait() (<-chan bool, error) {
	return nil, nil
}

func (r *fakePlugin) Listen(string) *entities.Broadcast[plugin_entities.SessionMessage] {
	return nil
}

func (r *fakePlugin) Write(string, []byte) {
}

func (r *fakePlugin) WaitStarted() <-chan bool {
	c := make(chan bool)
	close(c)
	return c
}

func (r *fakePlugin) WaitStopped() <-chan bool {
	c := make(chan bool)
	return c
}

func getRandomPluginRuntime() *fakePlugin {
	return &fakePlugin{
		PluginRuntime: plugin_entities.PluginRuntime{
			Config: plugin_entities.PluginDeclaration{
				PluginDeclarationWithoutAdvancedFields: plugin_entities.PluginDeclarationWithoutAdvancedFields{
					Name: uuid.New().String(),
					Label: plugin_entities.I18nObject{
						EnUS: "label",
					},
					Version:   "0.0.1",
					Type:      plugin_entities.PluginType,
					Author:    "Yeuoly",
					CreatedAt: time.Now(),
					Plugins: plugin_entities.PluginExtensions{
						Tools: []string{"test"},
					},
				},
			},
		},
	}
}

type fakeRemotePluginServer struct {
}

func (f *fakeRemotePluginServer) Launch() error {
	return nil
}

func (f *fakeRemotePluginServer) Next() bool {
	return false
}

func (f *fakeRemotePluginServer) Read() (plugin_entities.PluginFullDuplexLifetime, error) {
	return nil, nil
}

func (f *fakeRemotePluginServer) Stop() error {
	return nil
}

func (f *fakeRemotePluginServer) Wrap(fn func(plugin_entities.PluginFullDuplexLifetime)) {
	fn(getRandomPluginRuntime())
}

func TestRemotePluginWatcherPluginStoredToManager(t *testing.T) {
	config := &app.Config{}
	config.SetDefault()
	routine.InitPool(1024)
	oss := local.NewLocalStorage("./storage")
	pm := InitGlobalManager(oss, &app.Config{})
	pm.remotePluginServer = &fakeRemotePluginServer{}
	pm.startRemoteWatcher(config)

	time.Sleep(1 * time.Second)

	if pm.m.Len() != 1 {
		t.Fatalf("Expected 1 plugin, got %d", pm.m.Len())
	}
}
