package cluster

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/positive_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

type fakePlugin struct {
	plugin_entities.PluginRuntime
	positive_manager.PositivePluginRuntime
}

func (r *fakePlugin) InitEnvironment() error {
	return nil
}

func (r *fakePlugin) Checksum() string {
	return ""
}

func (r *fakePlugin) Identity() (string, error) {
	return "", nil
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

func getRandomPluginRuntime() fakePlugin {
	return fakePlugin{
		PluginRuntime: plugin_entities.PluginRuntime{
			Config: plugin_entities.PluginDeclaration{
				Name: uuid.New().String(),
				Label: plugin_entities.I18nObject{
					EnUS: "label",
				},
				Version:   "0.0.1",
				Type:      plugin_entities.PluginType,
				Author:    "Yeuoly",
				CreatedAt: time.Now(),
				Plugins:   []string{"test"},
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

	launchSimulationCluster(cluster, t)
	defer closeSimulationCluster(cluster, t)

	time.Sleep(time.Second * 1)

	// add plugin to the cluster
	err = cluster[0].RegisterPlugin(&plugin)
	if err != nil {
		t.Errorf("register plugin failed: %v", err)
		return
	}

	hashed_identity, err := plugin.HashedIdentity()
	if err != nil {
		t.Errorf("get plugin hashed identity failed: %v", err)
		return
	}

	nodes, err := cluster[0].FetchPluginAvailableNodesByHashedId(hashed_identity)
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
	nodes, err = cluster[0].FetchPluginAvailableNodesByHashedId(hashed_identity)
	if err != nil {
		t.Errorf("fetch plugin available nodes failed: %v", err)
		return
	}

	if len(nodes) != 0 {
		t.Errorf("plugin not stopped")
		return
	}
}

func TestPluginScheduleWhenMasterClusterShutdown(t *testing.T) {
	plugins := []fakePlugin{
		getRandomPluginRuntime(),
		getRandomPluginRuntime(),
	}

	cluster, err := createSimulationCluster(2)
	if err != nil {
		t.Errorf("create simulation cluster failed: %v", err)
		return
	}

	launchSimulationCluster(cluster, t)
	defer closeSimulationCluster(cluster, t)

	// add plugin to the cluster
	for i, plugin := range plugins {
		err = cluster[i].RegisterPlugin(&plugin)
		if err != nil {
			t.Errorf("register plugin failed: %v", err)
			return
		}
	}

	// wait for the plugin to be scheduled
	time.Sleep(time.Second * 1)

	// close master node and wait for new master to be elected
	master_idx := -1

	for i, node := range cluster {
		if node.IsMaster() {
			master_idx = i
			// close the master node
			node.Close()
			break
		}
	}

	if master_idx == -1 {
		t.Errorf("master node not found")
		return
	}

	// wait for the new master to be elected
	i := 0
	for ; i < 10; i++ {
		time.Sleep(time.Second * 1)
		found := false
		for i, node := range cluster {
			if node.IsMaster() && i != master_idx {
				found = true
				break
			}
		}

		if found {
			break
		}
	}

	if i == 10 {
		t.Errorf("master node is not elected")
		return
	}

	// check if plugins[master_idx] is removed
	hashed_identity, err := plugins[master_idx].HashedIdentity()
	if err != nil {
		t.Errorf("get plugin hashed identity failed: %v", err)
		return
	}

	ticker := time.NewTicker(time.Second)
	timeout := time.NewTimer(MASTER_GC_INTERVAL * 2)
	done := false
	for !done {
		select {
		case <-ticker.C:
			nodes, err := cluster[master_idx].FetchPluginAvailableNodesByHashedId(hashed_identity)
			if err != nil {
				t.Errorf("fetch plugin available nodes failed: %v", err)
				return
			}
			if len(nodes) == 0 {
				done = true
			}
		case <-timeout.C:
			t.Errorf("plugin not removed")
			return
		}
	}

	// check if plugins[1-master_idx] is still scheduled
	hashed_identity, err = plugins[1-master_idx].HashedIdentity()
	if err != nil {
		t.Errorf("get plugin hashed identity failed: %v", err)
		return
	}

	nodes, err := cluster[1-master_idx].FetchPluginAvailableNodesByHashedId(hashed_identity)
	if err != nil {
		t.Errorf("fetch plugin available nodes failed: %v", err)
		return
	}

	if len(nodes) != 1 {
		t.Errorf("plugin not scheduled")
		return
	}
}
