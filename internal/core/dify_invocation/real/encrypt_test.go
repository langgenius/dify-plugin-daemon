package real

import (
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/network"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func TestEncryptRequired(t *testing.T) {
	data := map[string]any{
		"key": "value",
	}

	payload := &dify_invocation.InvokeEncryptRequest{
		BaseInvokeDifyRequest: dify_invocation.BaseInvokeDifyRequest{
			TenantId: "123",
			UserId:   "456",
			Type:     dify_invocation.INVOKE_TYPE_ENCRYPT,
		},
		InvokeEncryptSchema: dify_invocation.InvokeEncryptSchema{
			Opt:       dify_invocation.ENCRYPT_OPT_ENCRYPT,
			Namespace: dify_invocation.ENCRYPT_NAMESPACE_ENDPOINT,
			Identity:  "test123",
			Data:      data,
			Config: []plugin_entities.ProviderConfig{
				{
					Name: "key",
					Type: plugin_entities.CONFIG_TYPE_SECRET_INPUT,
				},
			},
		},
	}

	if !payload.EncryptRequired(data) {
		t.Errorf("EncryptRequired should return true")
	}

	payload.Config = []plugin_entities.ProviderConfig{
		{
			Name: "key",
			Type: plugin_entities.CONFIG_TYPE_TEXT_INPUT,
		},
	}

	if payload.EncryptRequired(data) {
		t.Errorf("EncryptRequired should return false")
	}
}

func TestInvokeEncrypt(t *testing.T) {
	server := gin.Default()
	gin.SetMode(gin.ReleaseMode)

	port, err := network.GetRandomPort()
	if err != nil {
		t.Errorf("GetRandomPort failed: %v", err)
	}

	httpInvoked := false

	server.POST("/inner/api/invoke/encrypt", func(ctx *gin.Context) {
		data := make(map[string]any)
		if err := ctx.BindJSON(&data); err != nil {
			t.Errorf("BindJSON failed: %v", err)
		}

		if data["data"].(map[string]any)["key"] != "value" {
			t.Errorf("data[key] should be `value`, but got %v", data["data"].(map[string]any)["key"])
		}

		httpInvoked = true

		ctx.JSON(http.StatusOK, gin.H{
			"data": map[string]any{
				"data": map[string]any{
					"key": "encrypted",
				},
			},
		})
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: server,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	defer srv.Close()

	time.Sleep(1 * time.Second)

	i, err := NewDifyInvocationDaemon(NewDifyInvocationDaemonPayload{
		BaseUrl:      fmt.Sprintf("http://localhost:%d", port),
		CallingKey:   "test",
		WriteTimeout: 5000,
		ReadTimeout:  240000,
	})
	if err != nil {
		t.Errorf("InitDifyInvocationDaemon failed: %v", err)
		return
	}

	payload := &dify_invocation.InvokeEncryptRequest{
		BaseInvokeDifyRequest: dify_invocation.BaseInvokeDifyRequest{
			TenantId: "123",
			UserId:   "456",
			Type:     dify_invocation.INVOKE_TYPE_ENCRYPT,
		},
		InvokeEncryptSchema: dify_invocation.InvokeEncryptSchema{
			Opt:       dify_invocation.ENCRYPT_OPT_ENCRYPT,
			Namespace: dify_invocation.ENCRYPT_NAMESPACE_ENDPOINT,
			Identity:  "test123",
			Data:      map[string]any{"key": "value"},
			Config: []plugin_entities.ProviderConfig{
				{
					Name: "key",
					Type: plugin_entities.CONFIG_TYPE_SECRET_INPUT,
				},
			},
		},
	}

	if encrypted, err := i.InvokeEncrypt(payload); err != nil {
		t.Errorf("InvokeEncrypt failed: %v", err)
	} else {
		if encrypted["key"] != "encrypted" {
			t.Errorf("encrypted[key] should be `encrypted`, but got %v", encrypted["key"])
		}
	}

	if !httpInvoked {
		t.Errorf("http_invoked should be true")
	}
}
