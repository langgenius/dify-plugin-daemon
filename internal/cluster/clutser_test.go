package cluster

import (
	"testing"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

func createSimulationCluster(nums int) ([]*Cluster, error) {
	err := cache.InitRedisClient("0.0.0.0:6379", "difyai123456")
	if err != nil {
		return nil, err
	}

	result := make([]*Cluster, 0)
	for i := 0; i < nums; i++ {
		result = append(result, NewCluster(&app.Config{
			ServerPort: 12121,
		}))
	}

	log.SetShowLog(false)

	routine.InitPool(1024)

	return result, nil
}

func TestSingleClusterLifetime(t *testing.T) {
	clusters, err := createSimulationCluster(1)
	if err != nil {
		t.Errorf("create simulation cluster failed: %v", err)
		return
	}
	clusters[0].Launch()
	defer func() {
		// check if the cluster is closed
		_, err := cache.GetMapField[node](CLUSTER_STATUS_HASH_MAP_KEY, clusters[0].id)
		if err == nil {
			t.Errorf("cluster is not closed")
			return
		}
	}()
	defer clusters[0].Close()

	time.Sleep(time.Second * 1)

	_, err = cache.GetMapField[node](CLUSTER_STATUS_HASH_MAP_KEY, clusters[0].id)
	if err != nil {
		t.Errorf("get cluster status failed: %v", err)
		return
	}
}

func TestMultipleClusterLifetime(t *testing.T) {
	clusters, err := createSimulationCluster(3)
	if err != nil {
		t.Errorf("create simulation cluster failed: %v", err)
		return
	}

	for _, cluster := range clusters {
		cluster.Launch()
		defer func(cluster *Cluster) {
			cluster.Close()
			// wait for the cluster to close
			time.Sleep(time.Second * 1)
			// check if the cluster is closed
			_, err := cache.GetMapField[node](CLUSTER_STATUS_HASH_MAP_KEY, cluster.id)
			if err == nil {
				t.Errorf("cluster is not closed")
				return
			}
		}(cluster)
	}

	time.Sleep(time.Second * 1)

	has_master := false

	for _, cluster := range clusters {
		_, err := cache.GetMapField[node](CLUSTER_STATUS_HASH_MAP_KEY, cluster.id)
		if err != nil {
			t.Errorf("get cluster status failed: %v", err)
			return
		}

		if cluster.IsMaster() {
			if has_master {
				t.Errorf("multiple master")
				return
			} else {
				has_master = true
			}
		}
	}

	if !has_master {
		t.Errorf("no master")
	}
}

func TestClusterSubstituteMaster(t *testing.T) {
	clusters, err := createSimulationCluster(3)
	if err != nil {
		t.Errorf("create simulation cluster failed: %v", err)
		return
	}

	for _, cluster := range clusters {
		cluster.Launch()
		defer func(cluster *Cluster) {
			cluster.Close()
			// wait for the cluster to close
			time.Sleep(time.Second * 1)
			// check if the cluster is closed
			_, err := cache.GetMapField[node](CLUSTER_STATUS_HASH_MAP_KEY, cluster.id)
			if err == nil {
				t.Errorf("cluster is not closed")
				return
			}
		}(cluster)
	}

	time.Sleep(time.Second * 1)

	// close the master
	original_master_id := ""
	for _, cluster := range clusters {
		if cluster.IsMaster() {
			cluster.Close()
			original_master_id = cluster.id
			break
		}
	}
	if original_master_id == "" {
		t.Errorf("no master")
		return
	}

	time.Sleep(MASTER_LOCK_EXPIRED_TIME + time.Second)

	has_master := false

	for _, cluster := range clusters {
		if cluster.id == original_master_id {
			continue
		}
		_, err := cache.GetMapField[node](CLUSTER_STATUS_HASH_MAP_KEY, cluster.id)
		if err != nil {
			t.Errorf("get cluster status failed: %v", err)
			return
		}

		if cluster.IsMaster() {
			if has_master {
				t.Errorf("multiple substitute master")
				return
			} else {
				has_master = true
			}
		}
	}

	if !has_master {
		t.Errorf("no substitute master")
	}
}

func TestClusterAutoGCNoLongerActiveNode(t *testing.T) {
	clusters, err := createSimulationCluster(3)
	if err != nil {
		t.Errorf("create simulation cluster failed: %v", err)
		return
	}

	for _, cluster := range clusters {
		cluster.Launch()
		defer func(cluster *Cluster) {
			cluster.Close()
			// wait for the cluster to close
			time.Sleep(time.Second * 1)
			// check if the cluster is closed
			_, err := cache.GetMapField[node](CLUSTER_STATUS_HASH_MAP_KEY, cluster.id)
			if err == nil {
				t.Errorf("cluster is not closed")
				return
			}
		}(cluster)
	}

	time.Sleep(time.Second * 1)

	// randomly close a slave node to close
	slave_node_id := ""
	for _, cluster := range clusters {
		if !cluster.IsMaster() {
			slave_node_id = cluster.id
			cluster.Close()
			// wait for normal gc
			time.Sleep(time.Second * 1)
			// recover the node status
			if err := cluster.updateNodeStatus(); err != nil {
				t.Errorf("failed to recover the node status: %v", err)
				return
			}
			break
		}
	}

	if slave_node_id == "" {
		t.Errorf("no slave node")
		return
	}

	// wait for master gc task
	time.Sleep(MASTER_GC_INTERVAL * 2)

	_, err = cache.GetMapField[node](CLUSTER_STATUS_HASH_MAP_KEY, slave_node_id)
	if err == nil {
		t.Errorf("slave node is not collected by master gc automatically")
		return
	}
}
