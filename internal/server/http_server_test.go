package server

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/network"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
)

func TestServerHostBinding(t *testing.T) {
	tests := []struct {
		name           string
		host           string
		connectToHost  string
		wantStatusCode int
	}{
		{
			name:           "default host 0.0.0.0",
			host:           "0.0.0.0",
			connectToHost:  "127.0.0.1",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "localhost",
			host:           "127.0.0.1",
			connectToHost:  "127.0.0.1",
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port, err := network.GetRandomPort()
			if err != nil {
				t.Errorf("failed to get random port: %s", err.Error())
				return
			}

			config := &app.Config{
				ServerPort:          port,
				ServerHost:          tt.host,
				ServerKey:           "test-key",
				HealthApiLogEnabled: true,
				RoutinePoolSize:     100,
			}
			config.SetDefault()

			routine.InitPool(config.RoutinePoolSize)

			appInstance := &App{}
			cancel := appInstance.server(config)

			if cancel == nil {
				t.Errorf("failed to start server")
				return
			}
			defer cancel()

			time.Sleep(100 * time.Millisecond)

			client := &http.Client{Timeout: 5 * time.Second}
			url := "http://" + tt.connectToHost + ":" + strconv.Itoa(int(config.ServerPort)) + "/health/check"

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				t.Errorf("failed to create request: %s", err.Error())
				return
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Errorf("failed to send request: %s", err.Error())
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatusCode {
				t.Errorf("expected status %d, got %d", tt.wantStatusCode, resp.StatusCode)
			}
		})
	}
}
