package service

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

func InvokeTool(r *plugin_entities.InvokePluginRequest[plugin_entities.InvokeToolRequest], ctx *gin.Context) {
	// create session
	session := session_manager.NewSession(r.TenantId, r.UserId, parser.MarshalPluginIdentity(r.PluginName, r.PluginVersion))
	defer session.Close()

	writer := ctx.Writer
	writer.WriteHeader(200)
	writer.Header().Set("Content-Type", "text/event-stream")

	done := make(chan bool)

	write_data := func(data interface{}) {
		writer.WriteString("data: ")
		writer.Write([]byte(parser.MarshalJson(data)))
		writer.Write([]byte("\n\n"))
		writer.Flush()
	}

	plugin_daemon_response, err := plugin_daemon.InvokeTool(
		session,
		r.Data.ProviderName,
		r.Data.ToolName,
		r.Data.Parameters,
	)

	if err != nil {
		write_data(entities.NewErrorResponse(-500, err.Error()))
		close(done)
		return
	}

	routine.Submit(func() {
		for plugin_daemon_response.Next() {
			chunk, err := plugin_daemon_response.Read()
			if err != nil {
				break
			}
			write_data(entities.NewSuccessResponse(chunk))
		}
		close(done)
	})

	select {
	case <-writer.CloseNotify():
		plugin_daemon_response.Close()
	case <-done:
	}
}

func InvokeModel(r *plugin_entities.InvokePluginRequest[plugin_entities.InvokeModelRequest], ctx *gin.Context) {

}
