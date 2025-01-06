package local_runtime

import (
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
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

func (r *LocalPluginRuntime) Write(session_id string, data []byte) {
	writeToStdioHandler(r.ioIdentity, append(data, '\n'))
}
