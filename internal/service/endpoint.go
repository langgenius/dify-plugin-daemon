package service

import (
	"bytes"
	"context"
	"encoding/hex"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

func Endpoint(
	ctx *gin.Context,
	endpoint *models.Endpoint,
	path string,
) {
	req := ctx.Request.Clone(context.Background())
	req.URL.Path = path

	var buffer bytes.Buffer
	err := req.Write(&buffer)

	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
	}

	// fetch plugin
	manager := plugin_manager.GetGlobalPluginManager()
	runtime := manager.Get(endpoint.PluginID)
	if runtime == nil {
		ctx.JSON(404, gin.H{"error": "plugin not found"})
		return
	}

	session := session_manager.NewSession(
		endpoint.TenantID,
		"",
		endpoint.PluginID,
		ctx.GetString("cluster_id"),
		access_types.PLUGIN_ACCESS_TYPE_Endpoint,
		access_types.PLUGIN_ACCESS_ACTION_INVOKE_ENDPOINT,
		runtime.Configuration(),
	)
	defer session.Close()

	session.BindRuntime(runtime)

	status_code, headers, response, err := plugin_daemon.InvokeEndpoint(session, &requests.RequestInvokeEndpoint{
		RawHttpRequest: hex.EncodeToString(buffer.Bytes()),
	})
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer response.Close()

	done := make(chan bool)
	closed := new(int32)

	ctx.Status(status_code)
	for k, v := range *headers {
		if len(v) > 0 {
			ctx.Writer.Header().Set(k, v[0])
		}
	}

	close := func() {
		if atomic.CompareAndSwapInt32(closed, 0, 1) {
			close(done)
		}
	}
	defer close()

	routine.Submit(func() {
		defer close()
		for response.Next() {
			chunk, err := response.Read()
			if err != nil {
				ctx.JSON(500, gin.H{"error": err.Error()})
				return
			}
			ctx.Writer.Write(chunk)
			ctx.Writer.Flush()
		}
	})

	select {
	case <-ctx.Writer.CloseNotify():
	case <-done:
	case <-time.After(30 * time.Second):
		ctx.JSON(500, gin.H{"error": "killed by timeout"})
	}
}
