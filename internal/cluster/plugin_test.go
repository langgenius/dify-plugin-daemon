package cluster

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/basic_runtime"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/manifest_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

type fakePlugin struct {
	plugin_entities.PluginRuntime
	basic_runtime.BasicChecksum
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

func (r *fakePlugin) Listen(string) (*entities.Broadcast[plugin_entities.SessionMessage], error) {
	return nil, nil
}

func (r *fakePlugin) Write(string, access_types.PluginAccessAction, []byte) error {
	return nil
}

func getRandomPluginRuntime() fakePlugin {
	return fakePlugin{
		PluginRuntime: plugin_entities.PluginRuntime{
			Config: plugin_entities.PluginDeclaration{
				PluginDeclarationWithoutAdvancedFields: plugin_entities.PluginDeclarationWithoutAdvancedFields{
					Name: uuid.New().String(),
					Label: plugin_entities.I18nObject{
						EnUS: "label",
					},
					Version:   "0.0.1",
					Type:      manifest_entities.PluginType,
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

func TestPluginScheduleLifetime(t *testing.T) {
	plugin := getRandomPluginRuntime()
	cluster, err := createSimulationCluster(1)
	if err != nil {
		t.Errorf("create simulation cluster failed: %v", err)
		return
	}

	launchSimulationCluster(cluster)
	defer closeSimulationCluster(cluster, t)

	time.Sleep(time.Second * 1)

	// add plugin to the cluster
	err = cluster[0].RegisterPlugin(&plugin)
	if err != nil {
		t.Errorf("register plugin failed: %v", err)
		return
	}

	identity, err := plugin.Identity()
	if err != nil {
		t.Errorf("get plugin identity failed: %v", err)
		return
	}

	hashedIdentity := plugin_entities.HashedIdentity(identity.String())

	nodes, err := cluster[0].FetchPluginAvailableNodesByHashedId(hashedIdentity)
	if err != nil {
		t.Errorf("fetch plugin available nodes failed: %v", err)
		return
	}

	if len(nodes) != 1 {
		t.Errorf("plugin not scheduled")
		return
	}

	if nodes[0] != cluster[0].id {
		t.Errorf("plugin scheduled to wrong node")
		return
	}

	// trigger plugin stop
	plugin.Stop()

	// notify plugin has been stopped
	if err := cluster[0].UnregisterPlugin(&plugin); err != nil {
		t.Errorf("unregister plugin failed: %v", err)
		return
	}

	// wait for the plugin to stop
	time.Sleep(time.Second * 1)

	// check if the plugin is stopped
	nodes, err = cluster[0].FetchPluginAvailableNodesByHashedId(hashedIdentity)
	if err != nil {
		t.Errorf("fetch plugin available nodes failed: %v", err)
		return
	}

	if len(nodes) != 0 {
		t.Errorf("plugin not stopped")
		return
	}
}

func TestPluginRegisterIdempotent(t *testing.T) {
	plugin := getRandomPluginRuntime()
	cluster, err := createSimulationCluster(1)
	if err != nil {
		t.Errorf("create simulation cluster failed: %v", err)
		return
	}

	launchSimulationCluster(cluster)
	defer closeSimulationCluster(cluster, t)

	// wait for cluster to be ready
	time.Sleep(time.Second * 1)

	// first registration should succeed
	err = cluster[0].RegisterPlugin(&plugin)
	if err != nil {
		t.Errorf("first register plugin failed: %v", err)
		return
	}

	identity, err := plugin.Identity()
	if err != nil {
		t.Errorf("get plugin identity failed: %v", err)
		return
	}

	hashedIdentity := plugin_entities.HashedIdentity(identity.String())

	// wait for plugin to be scheduled
	time.Sleep(time.Second * 1)

	// verify plugin is registered
	nodes, err := cluster[0].FetchPluginAvailableNodesByHashedId(hashedIdentity)
	if err != nil {
		t.Errorf("fetch plugin available nodes failed: %v", err)
		return
	}

	if len(nodes) != 1 {
		t.Errorf("plugin not scheduled after first registration")
		return
	}

	// second registration with same identity should be idempotent (no error)
	err = cluster[0].RegisterPlugin(&plugin)
	if err != nil {
		t.Errorf("second register plugin failed (should be idempotent): %v", err)
		return
	}

	// verify plugin is still registered after second registration
	nodes, err = cluster[0].FetchPluginAvailableNodesByHashedId(hashedIdentity)
	if err != nil {
		t.Errorf("fetch plugin available nodes failed after second registration: %v", err)
		return
	}

	if len(nodes) != 1 {
		t.Errorf("plugin not available after second registration")
		return
	}

	// unregister the plugin
	if err := cluster[0].UnregisterPlugin(&plugin); err != nil {
		t.Errorf("unregister plugin failed: %v", err)
		return
	}

	// verify plugin is unregistered
	nodes, err = cluster[0].FetchPluginAvailableNodesByHashedId(hashedIdentity)
	if err != nil {
		t.Errorf("fetch plugin available nodes failed after unregister: %v", err)
		return
	}

	if len(nodes) != 0 {
		t.Errorf("plugin still available after unregister")
		return
	}

	// registration after unregister should succeed again
	err = cluster[0].RegisterPlugin(&plugin)
	if err != nil {
		t.Errorf("register after unregister failed: %v", err)
		return
	}

	// verify plugin is registered again
	nodes, err = cluster[0].FetchPluginAvailableNodesByHashedId(hashedIdentity)
	if err != nil {
		t.Errorf("fetch plugin available nodes failed after re-registration: %v", err)
		return
	}

	if len(nodes) != 1 {
		t.Errorf("plugin not available after re-registration")
		return
	}
}
