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
	err := cache.InitRedisClient("0.0.0.0:6379", "difyai123456", false, 0)
	if err != nil {
		return nil, err
	}

	result := make([]*Cluster, 0)
	for i := 0; i < nums; i++ {
		result = append(result, NewCluster(&app.Config{
			ServerPort: 12121,
		}, nil))
	}

	log.SetShowLog(false)

	routine.InitPool(1024)

	// delete master key
	if err := cache.Del(PREEMPTION_LOCK_KEY); err != nil {
		return nil, err
	}

	return result, nil
}

func launchSimulationCluster(clusters []*Cluster) {
	for _, cluster := range clusters {
		cluster.Launch()
	}
}

func closeSimulationCluster(clusters []*Cluster, t *testing.T) {
	for _, cluster := range clusters {
		cluster.Close()
		// wait for the cluster to close
		<-cluster.NotifyClusterStopped()
		// check if the cluster is closed
		_, err := cache.GetMapField[node](CLUSTER_STATUS_HASH_MAP_KEY, cluster.id)
		if err == nil {
			t.Errorf("cluster is not closed")
			return
		}
	}
}

func TestSingleClusterLifetime(t *testing.T) {
	clusters, err := createSimulationCluster(1)
	if err != nil {
		t.Errorf("create simulation cluster failed: %v", err)
		return
	}
	launchSimulationCluster(clusters)
	defer closeSimulationCluster(clusters, t)

	<-clusters[0].NotifyBecomeMaster()

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
	launchSimulationCluster(clusters)
	defer closeSimulationCluster(clusters, t)

	select {
	case <-clusters[0].NotifyBecomeMaster():
	case <-clusters[1].NotifyBecomeMaster():
	case <-clusters[2].NotifyBecomeMaster():
	}

	hasMaster := false

	for _, cluster := range clusters {
		_, err := cache.GetMapField[node](CLUSTER_STATUS_HASH_MAP_KEY, cluster.id)
		if err != nil {
			t.Errorf("get cluster status failed: %v", err)
			return
		}

		if cluster.IsMaster() {
			if hasMaster {
				t.Errorf("multiple master")
				return
			} else {
				hasMaster = true
			}
		}
	}

	if !hasMaster {
		t.Errorf("no master")
	}
}

func TestClusterSubstituteMaster(t *testing.T) {
	clusters, err := createSimulationCluster(3)
	if err != nil {
		t.Errorf("create simulation cluster failed: %v", err)
		return
	}
	launchSimulationCluster(clusters)
	defer closeSimulationCluster(clusters, t)

	select {
	case <-clusters[0].NotifyBecomeMaster():
	case <-clusters[1].NotifyBecomeMaster():
	case <-clusters[2].NotifyBecomeMaster():
	}

	// close the master
	originalMasterId := ""
	for _, cluster := range clusters {
		if cluster.IsMaster() {
			cluster.Close()
			originalMasterId = cluster.id
			break
		}
	}
	if originalMasterId == "" {
		t.Errorf("no master")
		return
	}

	time.Sleep(clusters[0].masterLockExpiredTime + time.Second)

	hasMaster := false

	for _, cluster := range clusters {
		if cluster.id == originalMasterId {
			continue
		}
		_, err := cache.GetMapField[node](CLUSTER_STATUS_HASH_MAP_KEY, cluster.id)
		if err != nil {
			t.Errorf("get cluster status failed: %v", err)
			return
		}

		if cluster.IsMaster() {
			if hasMaster {
				t.Errorf("multiple substitute master")
				return
			} else {
				hasMaster = true
			}
		}
	}

	if !hasMaster {
		t.Errorf("no substitute master")
	}
}

func TestClusterAutoGCNoLongerActiveNode(t *testing.T) {
	clusters, err := createSimulationCluster(3)
	if err != nil {
		t.Errorf("create simulation cluster failed: %v", err)
		return
	}
	launchSimulationCluster(clusters)
	defer closeSimulationCluster(clusters, t)

	select {
	case <-clusters[0].NotifyBecomeMaster():
	case <-clusters[1].NotifyBecomeMaster():
	case <-clusters[2].NotifyBecomeMaster():
	}

	// randomly close a slave node to close
	slaveNodeId := ""
	for _, cluster := range clusters {
		if !cluster.IsMaster() {
			slaveNodeId = cluster.id
			cluster.Close()
			// wait for the cluster to close
			<-cluster.NotifyClusterStopped()
			// recover the node status
			if err := cluster.updateNodeStatus(); err != nil {
				t.Errorf("failed to recover the node status: %v", err)
				return
			}
			break
		}
	}

	if slaveNodeId == "" {
		t.Errorf("no slave node")
		return
	}

	// wait for master gc task
	time.Sleep(clusters[0].nodeDisconnectedTimeout*2 + time.Second)

	_, err = cache.GetMapField[node](CLUSTER_STATUS_HASH_MAP_KEY, slaveNodeId)
	if err == nil {
		t.Errorf("slave node is not collected by master gc automatically")
		return
	}
}
