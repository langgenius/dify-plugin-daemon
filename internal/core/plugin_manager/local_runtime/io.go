package local_runtime

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func (r *LocalPluginRuntime) Listen(session_id string) *entities.Broadcast[plugin_entities.SessionMessage] {
	listener := entities.NewBroadcast[plugin_entities.SessionMessage]()
	listener.OnClose(func() {
		removeStdioHandlerListener(r.ioIdentity, session_id)
	})
	setupStdioEventListener(r.ioIdentity, session_id, func(b []byte) {
		// unmarshal the session message
		data, err := parser.UnmarshalJsonBytes[plugin_entities.SessionMessage](b)
		if err != nil {
			log.Error("unmarshal json failed: %s, failed to parse session message", err.Error())
			return
		}

		listener.Send(data)
	})
	return listener
}

func (r *LocalPluginRuntime) Write(session_id string, action access_types.PluginAccessAction, data []byte) {
	writeToStdioHandler(r.ioIdentity, append(data, '\n'))
}
