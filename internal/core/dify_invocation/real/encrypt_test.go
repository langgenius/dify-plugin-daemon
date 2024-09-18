package real

import (
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/network"
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
			Config: map[string]plugin_entities.ProviderConfig{
				"key": {
					Name: "key",
					Type: plugin_entities.CONFIG_TYPE_SECRET_INPUT,
				},
			},
		},
	}

	if !payload.EncryptRequired(data) {
		t.Errorf("EncryptRequired should return true")
	}

	payload.Config["key"] = plugin_entities.ProviderConfig{
		Name: "key",
		Type: plugin_entities.CONFIG_TYPE_TEXT_INPUT,
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

	http_invoked := false

	server.POST("/inner/api/invoke/encrypt", func(ctx *gin.Context) {
		data := make(map[string]any)
		if err := ctx.BindJSON(&data); err != nil {
			t.Errorf("BindJSON failed: %v", err)
		}

		if data["data"].(map[string]any)["key"] != "value" {
			t.Errorf("data[key] should be `value`, but got %v", data["data"].(map[string]any)["key"])
		}

		http_invoked = true

		ctx.JSON(http.StatusOK, gin.H{
			"key": "encrypted",
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

	i, err := InitDifyInvocationDaemon(fmt.Sprintf("http://localhost:%d", port), "test")
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
			Config: map[string]plugin_entities.ProviderConfig{
				"key": {
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

	if !http_invoked {
		t.Errorf("http_invoked should be true")
	}
}
