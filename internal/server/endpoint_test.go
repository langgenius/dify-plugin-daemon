package server

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/network"
)

func TestEndpointParams(t *testing.T) {
	port, err := network.GetRandomPort()
	if err != nil {
		t.Errorf("failed to get random port: %s", err.Error())
		return
	}

	global_hook_id := ""
	global_hook_path := ""

	handler := func(ctx *gin.Context, hook_id string, path string) {
		global_hook_id = hook_id
		global_hook_path = path
	}

	app_pointer := &App{
		endpoint_handler: handler,
	}
	cancel := app_pointer.server(&app.Config{
		ServerPort:            port,
		PluginEndpointEnabled: true,
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

	if global_hook_id != "1111" {
		t.Errorf("hook id not match: %s", global_hook_id)
		return
	}

	if global_hook_path != "/v1/chat/completions" {
		t.Errorf("hook path not match: %s", global_hook_path)
		return
	}
}
