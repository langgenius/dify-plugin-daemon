package cluster

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/network"
)

func createSimulationHealthCheckSever() (uint16, error) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.GET("/health/check", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	port, err := network.GetRandomPort()
	if err != nil {
		return 0, err
	}

	go func() {
		router.Run(fmt.Sprintf(":%d", port))
	}()

	return uint16(port), nil
}

func TestVoteAddresses(t *testing.T) {
	// create a health check server
	port, err := createSimulationHealthCheckSever()
	if err != nil {
		t.Errorf("create simulation health check server failed: %v", err)
		return
	}

	cluster, err := createSimulationCluster(2)
	if err != nil {
		t.Errorf("create simulation cluster failed: %v", err)
		return
	}

	for _, node := range cluster {
		node.port = port
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	// wait for all voting processes complete
	for _, node := range cluster {
		node := node
		go func() {
			defer wg.Done()
			<-node.NotifyVotingCompleted()
		}()
	}

	launchSimulationCluster(cluster)
	defer closeSimulationCluster(cluster, t)

	// wait for all nodes to be ready
	wg.Wait()

	// wait for all addresses to be voted
	time.Sleep(time.Second)

	for _, node := range cluster {
		nodes, err := node.GetNodes()
		if err != nil {
			t.Errorf("get nodes failed: %v", err)
			return
		}

		for _, node := range nodes {
			for _, ip := range node.Addresses {
				if len(ip.Votes) == 0 {
					t.Errorf("vote for ip %s failed", ip.Ip)
					return
				}
			}
		}
	}
}
