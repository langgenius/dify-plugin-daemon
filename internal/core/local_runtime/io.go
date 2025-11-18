package local_runtime

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
)

func (r *LocalPluginRuntime) Listen(sessionId string) (
	*entities.Broadcast[plugin_entities.SessionMessage], error,
) {
	// pick the instance with lowest load
	instance, err := r.pickLowestLoadInstance()
	if err != nil {
		return nil, err
	}

	// keep the mapping between sessionId and instance
	r.sessionToInstanceMap.Store(sessionId, instance)

	// setup listener to handle session message from plugin
	listener := entities.NewCallbackHandler[plugin_entities.SessionMessage]()
	listener.OnClose(func() {
		instance.removeStdioHandlerListener(sessionId)
		r.sessionToInstanceMap.Delete(sessionId)
	})

	instance.setupStdioEventListener(sessionId, func(b []byte) {
		data, err := parser.UnmarshalJsonBytes[plugin_entities.SessionMessage](b)
		if err != nil {
			log.Error("unmarshal json failed: %s, failed to parse session message", err.Error())
			return
		}

		listener.Send(data)
	})

	return listener, nil
}

func (r *LocalPluginRuntime) Write(
	sessionId string,
	action access_types.PluginAccessAction,
	data []byte,
) error {
	// get the instance from the mapping
	instance, ok := r.sessionToInstanceMap.Load(sessionId)
	if !ok {
		return ErrSessionNotFound
	}

	// write to the instance
	return instance.Write(append(data, '\n'))
}
