package cluster

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/endpoint_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/network"
	"github.com/stretchr/testify/assert"
)

// Helper function to create test requests
func createTestRequest(url string) *http.Request {
	req, _ := http.NewRequest("GET", url, nil)
	return req
}

// MockCluster extends Cluster for testing
type MockCluster struct {
	*Cluster
	id   string
	port uint16
}

func (m *MockCluster) ID() string {
	return m.id
}

type SimulationCheckServer struct {
	http.Server

	port uint16
}

func createSimulationSevers(nums int, registerCallback func(i int, c *gin.Engine)) ([]*SimulationCheckServer, error) {
	gin.SetMode(gin.ReleaseMode)
	engines := make([]*gin.Engine, nums)
	servers := make([]*SimulationCheckServer, nums)
	for i := 0; i < nums; i++ {
		engines[i] = gin.Default()
		registerCallback(i, engines[i])
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

	node1RecvReqs := make(chan struct{})
	node1RecvCorrectReqs := make(chan struct{})
	defer close(node1RecvReqs)
	defer close(node1RecvCorrectReqs)

	// create 2 simulated servers
	servers, err := createSimulationSevers(2, func(i int, c *gin.Engine) {
		c.GET("/plugin/invoke/tool", func(c *gin.Context) {
			if i == 0 {
				// redirect to node 1
				statusCode, headers, reader, err := cluster[i].RedirectRequest(cluster[1].id, c.Request)
				if err != nil {
					c.String(http.StatusInternalServerError, err.Error())
					return
				}
				c.Status(statusCode)
				for k, v := range headers {
					for _, vv := range v {
						c.Header(k, vv)
					}
				}
				defer reader.Close()
				io.Copy(c.Writer, reader)
			} else {
				c.String(http.StatusOK, "ok")
				node1RecvReqs <- struct{}{}
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
				node1RecvCorrectReqs <- struct{}{}
			}
		}
	}()

	// check if node 1 received the request
	recvCount := 0
	correctCount := 0
	for {
		select {
		case <-node1RecvReqs:
			recvCount++
		case <-node1RecvCorrectReqs:
			correctCount++
			if correctCount == 10 {
				return
			}
		case <-time.After(5 * time.Second):
			t.Fatal("node 1 did not receive correct requests")
		}
	}
}

func TestRedirectTrafficWithQueryParams(t *testing.T) {
	request, err := http.NewRequest("GET", "http://localhost:8080/plugin/invoke/tool?a=1&b=2", nil)
	if err != nil {
		t.Fatal(err)
	}

	request.Header.Set(endpoint_entities.HeaderXOriginalHost, "localhost:8080")

	ip := address{
		Ip:   "127.0.0.1",
		Port: 8080,
	}

	redirectedRequest := constructRedirectUrl(ip, request)
	if redirectedRequest != "http://127.0.0.1:8080/plugin/invoke/tool?a=1&b=2" {
		t.Fatal("redirected request is not correct")
	}
}

func TestRedirectTrafficWithOutQueryParams(t *testing.T) {
	request, err := http.NewRequest("GET", "http://localhost:8080/plugin/invoke/tool", nil)
	if err != nil {
		t.Fatal(err)
	}

	request.Header.Set(endpoint_entities.HeaderXOriginalHost, "localhost:8080")

	ip := address{
		Ip:   "127.0.0.1",
		Port: 8080,
	}

	redirectedRequest := constructRedirectUrl(ip, request)
	if redirectedRequest != "http://127.0.0.1:8080/plugin/invoke/tool" {
		t.Fatal("redirected request is not correct")
	}
}

func TestRedirectTrafficWithPathStyle(t *testing.T) {
	payload := `{"a": "1", "b": "2"}`

	server := gin.Default()
	server.POST("/plugin/invoke/tool", func(c *gin.Context) {
		content, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		c.String(http.StatusOK, string(content))
	})

	port, err := network.GetRandomPort()
	if err != nil {
		t.Fatal(err)
	}

	srv := &SimulationCheckServer{
		Server: http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: server,
		},
		port: port,
	}

	go func() {
		srv.ListenAndServe()
	}()

	// wait for server to be ready
	time.Sleep(3 * time.Second)

	defer srv.Shutdown(context.Background())

	request, err := http.NewRequest("POST", "http://localhost:8080/plugin/invoke/tool", strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}

	// redirect to srv
	statusCode, _, reader, err := redirectRequestToIp(address{
		Ip:   "127.0.0.1",
		Port: port,
	}, request)
	if err != nil {
		t.Fatal(err)
	}

	content, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}

	if statusCode != http.StatusOK {
		t.Fatal("status code is not ok")
	}

	if string(content) != payload {
		t.Fatal("content is not correct")
	}
}

// Tests for localhost redirection using the generic constructor
func TestConstructRedirectUrlLocalhost(t *testing.T) {
	tests := []struct {
		name     string
		port     uint16
		request  *http.Request
		expected string
	}{
		{
			name:     "basic localhost URL",
			port:     5002,
			request:  createTestRequest("/plugin/test"),
			expected: "http://localhost:5002/plugin/test",
		},
		{
			name:     "localhost URL with query parameters",
			port:     8080,
			request:  createTestRequest("/api/v1/endpoint?param1=value1&param2=value2"),
			expected: "http://localhost:8080/api/v1/endpoint?param1=value1&param2=value2",
		},
		{
			name:     "localhost URL with complex path",
			port:     3000,
			request:  createTestRequest("/plugin/a5df51ca-fba9-4170-8369-4ae0eff4f543/dispatch/model/schema"),
			expected: "http://localhost:3000/plugin/a5df51ca-fba9-4170-8369-4ae0eff4f543/dispatch/model/schema",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr := address{Ip: "localhost", Port: tt.port}
			result := constructRedirectUrl(addr, tt.request)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRedirectRequestToLocalhostUsingGeneric(t *testing.T) {
	// Create a test server to simulate the local endpoint
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("local response"))
	}))
	defer testServer.Close()

	// Test with a request that will fail (no server on localhost:5002)
	req := httptest.NewRequest("GET", "/test", nil)
	statusCode, header, body, err := redirectRequestToIp(address{Ip: "localhost", Port: 5002}, req)

	// Should fail since there's no actual server on localhost:5002
	assert.Error(t, err)
	assert.Equal(t, 0, statusCode)
	assert.Nil(t, header)
	assert.Nil(t, body)
}

func TestRedirectRequestToLocalhostWithActualServerUsingGeneric(t *testing.T) {
	// Create a test server on localhost
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/test", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer testServer.Close()

	// Extract the port from the test server URL
	parts := strings.Split(testServer.URL, ":")
	port := parts[len(parts)-1]
	portNum := uint16(0)
	fmt.Sscanf(port, "%d", &portNum)

	// Create request
	req := httptest.NewRequest("GET", "/test", nil)

	// This should work since we have an actual server
	statusCode, header, body, err := redirectRequestToIp(address{Ip: "localhost", Port: portNum}, req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
	assert.NotNil(t, header)
	assert.NotNil(t, body)

	// Read response body
	content, err := io.ReadAll(body)
	assert.NoError(t, err)
	assert.Equal(t, "success", string(content))

	// Close body
	body.Close()
}

func BenchmarkConstructRedirectUrlLocalhost(b *testing.B) {
	req := httptest.NewRequest("GET", "/plugin/test?param=value", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		constructRedirectUrl(address{Ip: "localhost", Port: 5002}, req)
	}
}

func TestClusterRedirectRequestToCurrentNode(t *testing.T) {
	// Create a mock cluster
	config := &app.Config{
		ServerPort: 5002,
	}

	cluster := &MockCluster{
		Cluster: NewCluster(config),
		id:      "test-node-id",
		port:    5002,
	}

	// Test redirect to current node (should use localhost)
	req := httptest.NewRequest("GET", "/plugin/test", nil)
	statusCode, header, body, err := cluster.RedirectRequest("test-node-id", req)

	// Should fail since no actual server on localhost:5002, but should attempt localhost
	assert.Error(t, err)
	assert.Equal(t, 0, statusCode)
	assert.Nil(t, header)
	assert.Nil(t, body)
}

func TestClusterRedirectRequestToUnknownNode(t *testing.T) {
	// Create a mock cluster
	config := &app.Config{
		ServerPort: 5002,
	}

	cluster := &MockCluster{
		Cluster: NewCluster(config),
		id:      "test-node-id",
		port:    5002,
	}

	// Test redirect to unknown node
	req := httptest.NewRequest("GET", "/plugin/test", nil)
	statusCode, header, body, err := cluster.RedirectRequest("unknown-node-id", req)

	// Should fail with "node not found" error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "node not found")
	assert.Equal(t, 0, statusCode)
	assert.Nil(t, header)
	assert.Nil(t, body)
}

func TestRedirectRequestWithTimeout(t *testing.T) {
	// Test that redirect requests have proper timeout
	ip := address{
		Ip:   "192.168.255.254", // Non-routable IP
		Port: 5002,
	}

	req := httptest.NewRequest("GET", "/test", nil)
	start := time.Now()

	statusCode, header, body, err := redirectRequestToIp(ip, req)

	elapsed := time.Since(start)

	// Should fail quickly due to timeout
	assert.Error(t, err)
	assert.Equal(t, 0, statusCode)
	assert.Nil(t, header)
	assert.Nil(t, body)
	assert.Less(t, elapsed, 15*time.Second) // Should timeout within 10 seconds + some buffer
}

// Benchmark tests
func BenchmarkConstructRedirectUrl(b *testing.B) {
	ip := address{Ip: "192.168.1.100", Port: 5002}
	req := httptest.NewRequest("GET", "/plugin/test?param=value", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		constructRedirectUrl(ip, req)
	}
}
