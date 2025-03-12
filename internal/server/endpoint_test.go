package server

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/network"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

func TestEndpointParams(t *testing.T) {
	port, err := network.GetRandomPort()
	if err != nil {
		t.Errorf("failed to get random port: %s", err.Error())
		return
	}

	globalHookId := ""
	globalHookPath := ""

	handler := func(ctx *gin.Context, hook_id string, maxExecutionTime time.Duration, path string) {
		globalHookId = hook_id
		globalHookPath = path
	}

	appPointer := &App{
		endpointHandler: handler,
	}
	cancel := appPointer.server(&app.Config{
		ServerPort:            port,
		PluginEndpointEnabled: parser.ToPtr(true),
		HealthApiLogEnabled:   parser.ToPtr(true),
	})
	defer cancel()

	// test endpoint params
	client := &http.Client{}
	req, err := http.NewRequest("POST", "http://localhost:"+strconv.Itoa(int(port))+"/e/1111/v1/chat/completions", nil)
	if err != nil {
		t.Errorf("failed to create request: %s", err.Error())
		return
	}
	_, err = client.Do(req)
	if err != nil {
		t.Errorf("failed to send request: %s", err.Error())
		return
	}

	if globalHookId != "1111" {
		t.Errorf("hook id not match: %s", globalHookId)
		return
	}

	if globalHookPath != "/v1/chat/completions" {
		t.Errorf("hook path not match: %s", globalHookPath)
		return
	}
}
