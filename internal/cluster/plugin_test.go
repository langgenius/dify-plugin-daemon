package cluster

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/access_types"
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

func (r *fakePlugin) Listen(string) *entities.Broadcast[plugin_entities.SessionMessage] {
	return nil
}

func (r *fakePlugin) Write(string, access_types.PluginAccessAction, []byte) {
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
	plugin.TriggerStop()

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

// TODO: I need to implement this test, now it's randomly working
// func TestPluginScheduleWhenMasterClusterShutdown(t *testing.T) {
// 	plugins := []fakePlugin{
// 		getRandomPluginRuntime(),
// 		getRandomPluginRuntime(),
// 	}

// 	cluster, err := createSimulationCluster(2)
// 	if err != nil {
// 		t.Errorf("create simulation cluster failed: %v", err)
// 		return
// 	}

// 	// set master gc interval to 1 second
// 	for _, node := range cluster {
// 		node.nodeDisconnectedTimeout = time.Second * 2
// 		node.masterGcInterval = time.Second * 1
// 		node.pluginSchedulerInterval = time.Second * 1
// 		node.pluginSchedulerTickerInterval = time.Second * 1
// 		node.updateNodeStatusInterval = time.Second * 1
// 		node.pluginDeactivatedTimeout = time.Second * 2
// 		node.showLog = true
// 	}

// 	launchSimulationCluster(cluster)
// 	defer closeSimulationCluster(cluster, t)

// 	// add plugin to the cluster
// 	for i, plugin := range plugins {
// 		err = cluster[i].RegisterPlugin(&plugin)
// 		if err != nil {
// 			t.Errorf("register plugin failed: %v", err)
// 			return
// 		}
// 	}

// 	// wait for the plugin to be scheduled
// 	time.Sleep(time.Second * 1)

// 	// close master node and wait for new master to be elected
// 	masterIdx := -1

// 	for i, node := range cluster {
// 		if node.IsMaster() {
// 			masterIdx = i
// 			// close the master node
// 			node.Close()
// 			break
// 		}
// 	}

// 	if masterIdx == -1 {
// 		t.Errorf("master node not found")
// 		return
// 	}

// 	// wait for the new master to be elected
// 	i := 0
// 	for ; i < 10; i++ {
// 		time.Sleep(time.Second * 1)
// 		found := false
// 		for i, node := range cluster {
// 			if node.IsMaster() && i != masterIdx {
// 				found = true
// 				break
// 			}
// 		}

// 		if found {
// 			break
// 		}
// 	}

// 	if i == 10 {
// 		t.Errorf("master node is not elected")
// 		return
// 	}

// 	// check if plugins[master_idx] is removed
// 	identity, err := plugins[masterIdx].Identity()
// 	if err != nil {
// 		t.Errorf("get plugin identity failed: %v", err)
// 		return
// 	}

// 	hashedIdentity := plugin_entities.HashedIdentity(identity.String())

// 	ticker := time.NewTicker(time.Second)
// 	timeout := time.NewTimer(time.Second * 20)
// 	done := false
// 	for !done {
// 		select {
// 		case <-ticker.C:
// 			nodes, err := cluster[masterIdx].FetchPluginAvailableNodesByHashedId(hashedIdentity)
// 			if err != nil {
// 				t.Errorf("fetch plugin available nodes failed: %v", err)
// 				return
// 			}
// 			if len(nodes) == 0 {
// 				done = true
// 			}
// 		case <-timeout.C:
// 			t.Errorf("plugin not removed")
// 			return
// 		}
// 	}

// 	// check if plugins[1-master_idx] is still scheduled
// 	identity, err = plugins[1-masterIdx].Identity()
// 	if err != nil {
// 		t.Errorf("get plugin identity failed: %v", err)
// 		return
// 	}

// 	hashedIdentity = plugin_entities.HashedIdentity(identity.String())

// 	nodes, err := cluster[1-masterIdx].FetchPluginAvailableNodesByHashedId(hashedIdentity)
// 	if err != nil {
// 		t.Errorf("fetch plugin available nodes failed: %v", err)
// 		return
// 	}

// 	if len(nodes) != 1 {
// 		t.Errorf("plugin not scheduled")
// 		return
// 	}
// }
