package cluster

import (
	"fmt"
	"net"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
)

func createSimulationHealthCheckSever() (uint16, error) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.GET("/health/check", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, fmt.Errorf("failed to get a random port: %s", err.Error())
	}
	listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port

	go func() {
		router.Run(fmt.Sprintf(":%d", port))
	}()

	return uint16(port), nil
}

func TestVoteIps(t *testing.T) {
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

	launchSimulationCluster(cluster, t)
	defer closeSimulationCluster(cluster, t)

	// wait for all nodes to be ready
	wg.Wait()
}
