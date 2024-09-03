package cluster

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/network"
)

type SimulationCheckServer struct {
	http.Server

	port uint16
}

func createSimulationSevers(nums int, register_callback func(i int, c *gin.Engine)) ([]*SimulationCheckServer, error) {
	gin.SetMode(gin.ReleaseMode)
	engines := make([]*gin.Engine, nums)
	servers := make([]*SimulationCheckServer, nums)
	for i := 0; i < nums; i++ {
		engines[i] = gin.Default()
		register_callback(i, engines[i])
	}

	// get random port
	ports := make([]uint16, nums)
	for i := 0; i < nums; i++ {
		port, err := network.GetRandomPort()
		if err != nil {
			return nil, err
		}
		ports[i] = port
	}

	for i := 0; i < nums; i++ {
		srv := &SimulationCheckServer{
			Server: http.Server{
				Addr:    fmt.Sprintf(":%d", ports[i]),
				Handler: engines[i],
			},
			port: ports[i],
		}
		servers[i] = srv

		go func(i int) {
			srv.ListenAndServe()
		}(i)
	}

	return servers, nil
}

func closeSimulationHealthCheckSevers(servers []*SimulationCheckServer) {
	for _, server := range servers {
		server.Shutdown(context.Background())
	}
}

func TestRedirectTraffic(t *testing.T) {
	clearClusterState()

	// create 2 nodes cluster
	cluster, err := createSimulationCluster(2)
	if err != nil {
		t.Fatal(err)
	}

	// wait for voting
	wg := sync.WaitGroup{}
	wg.Add(len(cluster))
	// wait for all voting processes complete
	for _, node := range cluster {
		node := node
		go func() {
			defer wg.Done()
			<-node.NotifyVotingCompleted()
		}()
	}

	node1_recv_reqs := make(chan struct{})
	node1_recv_correct_reqs := make(chan struct{})
	defer close(node1_recv_reqs)
	defer close(node1_recv_correct_reqs)

	// create 2 simulated servers
	servers, err := createSimulationSevers(2, func(i int, c *gin.Engine) {
		c.GET("/plugin/invoke/tool", func(c *gin.Context) {
			if i == 0 {
				// redirect to node 1
				status_code, headers, reader, err := cluster[i].RedirectRequest(cluster[1].id, c.Request)
				if err != nil {
					c.String(http.StatusInternalServerError, err.Error())
					return
				}
				c.Status(status_code)
				for k, v := range headers {
					for _, vv := range v {
						c.Header(k, vv)
					}
				}
				io.Copy(c.Writer, reader)
			} else {
				c.String(http.StatusOK, "ok")
				node1_recv_reqs <- struct{}{}
			}
		})
		c.GET("/health/check", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
			})
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	defer closeSimulationHealthCheckSevers(servers)

	// change port
	for i, node := range cluster {
		node.port = servers[i].port
	}

	// launch cluster
	launchSimulationCluster(cluster)
	defer closeSimulationCluster(cluster, t)

	// wait for all nodes to be ready
	wg.Wait()

	// wait for node status to by synchronized
	wg = sync.WaitGroup{}
	wg.Add(len(cluster))
	// wait for all voting processes complete
	for _, node := range cluster {
		node := node
		go func() {
			defer wg.Done()
			<-node.NotifyNodeUpdateCompleted()
		}()
	}
	wg.Wait()

	// request to node 0
	go func() {
		for i := 0; i < 10; i++ {
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/plugin/invoke/tool", servers[0].port))
			if err != nil {
				t.Error(err)
			}
			content, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Error(err)
			}
			if string(content) == "ok" {
				node1_recv_correct_reqs <- struct{}{}
			}
		}
	}()

	// check if node 1 received the request
	recv_count := 0
	correct_count := 0
	for {
		select {
		case <-node1_recv_reqs:
			recv_count++
		case <-node1_recv_correct_reqs:
			correct_count++
			if correct_count == 10 {
				return
			}
		case <-time.After(5 * time.Second):
			t.Fatal("node 1 did not receive correct requests")
		}
	}
}
