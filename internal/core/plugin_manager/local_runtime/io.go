package local_runtime

import (
	"fmt"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func (r *LocalPluginRuntime) Listen(sessionId string) (*entities.Broadcast[plugin_entities.SessionMessage], error) {
	listener := entities.NewBroadcast[plugin_entities.SessionMessage]()
	holder, err := r.matchStdioHolder(sessionId)
	if err != nil {
		return nil, err
	}
	listener.OnClose(func() {
		holder.removeStdioHandlerListener(sessionId)
	})
	holder.setupStdioEventListener(sessionId, func(b []byte) {
		// unmarshal the session message
		data, err := parser.UnmarshalJsonBytes[plugin_entities.SessionMessage](b)
		if err != nil {
			log.Error("unmarshal json failed: %s, failed to parse session message", err.Error())
			return
		}

		listener.Send(data)

		// received event from plugin, it has been verified to work properly
		r.stage = launchStageVerifiedWorking
	})

	return listener, nil
}

func (r *LocalPluginRuntime) matchStdioHolder(sessionId string) (*pluginInstance, error) {
	r.stdioHolderLock.Lock()
	defer r.stdioHolderLock.Unlock()

	key, ok := r.sessionToStdioHolder[sessionId]
	if ok {
		// fetch the stdio holder
		for _, holder := range r.stdioHolders {
			if holder.id == key.instanceId {
				return holder, nil
			}
		}

		return nil, fmt.Errorf("stdio holder for session %s not found, the plugin instance may dead", sessionId)
	}

	if len(r.stdioHolders) == 0 {
		return nil, fmt.Errorf("no stdio holder found, please wait for the plugin to start")
	}

	var holder *pluginInstance

	if r.autoScale {
		holder = r.getLowestLoadStdioHolder()
	} else {
		holder = r.stdioHolders[0]
	}

	// add the session to the stdio holder
	r.sessionToStdioHolder[sessionId] = &stdioHolderKey{
		instanceId: holder.id,
		attachedAt: time.Now(),
	}

	return holder, nil
}

func (r *LocalPluginRuntime) Write(sessionId string, action access_types.PluginAccessAction, data []byte) error {
	holder, err := r.matchStdioHolder(sessionId)
	if err != nil {
		return err
	}

	return holder.write(append(data, '\n'))
}
