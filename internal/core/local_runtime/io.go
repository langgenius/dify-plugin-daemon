package local_runtime

import (
	"errors"
	"io"
	"os"
	"slices"
	"strings"
	"syscall"

	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
)

func IsInstanceDeadErr(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, ErrInstanceStopped) ||
		errors.Is(err, os.ErrClosed) ||
		errors.Is(err, syscall.EPIPE) ||
		errors.Is(err, io.ErrClosedPipe) ||
		errors.Is(err, io.ErrShortWrite) {
		return true
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "file already closed") || strings.Contains(msg, "broken pipe")
}

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
			log.Error("unmarshal json failed, failed to parse session message", "error", err)
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
	err := instance.Write(append(data, '\n'))
	if err == nil {
		return nil
	}

	if IsInstanceDeadErr(err) {
		instance.Stop()

		r.instanceLocker.Lock()
		before := len(r.instances)
		r.instances = slices.DeleteFunc(r.instances, func(item *PluginInstance) bool {
			return item.instanceId == instance.instanceId
		})
		after := len(r.instances)
		if after == 0 {
			r.SetRestarting()
		}
		r.instanceLocker.Unlock()

		r.sessionToInstanceMap.Delete(sessionId)
		r.nudgeSchedule()

		log.Warn(
			"evict dead local plugin instance after write failure",
			"plugin", r.Config.Identity(),
			"instance", instance.ID()[:8],
			"session", sessionId,
			"action", action,
			"instances_before", before,
			"instances_after", after,
			"error", err,
		)
	}

	return err
}
