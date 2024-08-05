package server

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/network"
)

func TestWebhookParams(t *testing.T) {
	port, err := network.GetRandomPort()
	if err != nil {
		t.Errorf("failed to get random port: %s", err.Error())
		return
	}

	global_hook_id := ""
	global_hook_path := ""

	handler := func(hook_id string, path string) {
		global_hook_id = hook_id
		global_hook_path = path
	}

	app_pointer := &App{
		webhook_handler: handler,
	}
	cancel := app_pointer.server(&app.Config{
		ServerPort:           port,
		PluginWebhookEnabled: true,
	})
	defer cancel()

	// test webhook params
	client := &http.Client{}
	req, err := http.NewRequest("POST", "http://localhost:"+strconv.Itoa(int(port))+"/webhook/1111/v1/chat/completions", nil)
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
