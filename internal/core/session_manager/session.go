package session_manager

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

var (
	sessions     map[string]*Session = map[string]*Session{}
	session_lock sync.RWMutex
)

// session need to implement the backwards_invocation.BackwardsInvocationWriter interface
type Session struct {
	ID      string                                 `json:"id"`
	runtime plugin_entities.PluginRuntimeInterface `json:"-"`

	TenantID       string                             `json:"tenant_id"`
	UserID         string                             `json:"user_id"`
	PluginIdentity string                             `json:"plugin_identity"`
	ClusterID      string                             `json:"cluster_id"`
	InvokeFrom     access_types.PluginAccessType      `json:"invoke_from"`
	Action         access_types.PluginAccessAction    `json:"action"`
	Declaration    *plugin_entities.PluginDeclaration `json:"declaration"`
}

func sessionKey(id string) string {
	return fmt.Sprintf("session_info:%s", id)
}

func NewSession(
	tenant_id string,
	user_id string,
	plugin_identity string,
	cluster_id string,
	invoke_from access_types.PluginAccessType,
	action access_types.PluginAccessAction,
	declaration *plugin_entities.PluginDeclaration,
) *Session {
	s := &Session{
		ID:             uuid.New().String(),
		TenantID:       tenant_id,
		UserID:         user_id,
		PluginIdentity: plugin_identity,
		ClusterID:      cluster_id,
		InvokeFrom:     invoke_from,
		Action:         action,
		Declaration:    declaration,
	}

	session_lock.Lock()
	sessions[s.ID] = s
	session_lock.Unlock()

	if err := cache.Store(sessionKey(s.ID), s, time.Minute*30); err != nil {
		log.Error("set session info to cache failed, %s", err)
	}

	return s
}

func GetSession(id string) *Session {
	session_lock.RLock()
	defer session_lock.RUnlock()

	return sessions[id]
}

func DeleteSession(id string) {
	session_lock.Lock()
	delete(sessions, id)
	session_lock.Unlock()

	if err := cache.Del(sessionKey(id)); err != nil {
		log.Error("delete session info from cache failed, %s", err)
	}
}

func (s *Session) Close() {
	DeleteSession(s.ID)
}

func (s *Session) BindRuntime(runtime plugin_entities.PluginRuntimeInterface) {
	s.runtime = runtime
}

func (s *Session) Runtime() plugin_entities.PluginRuntimeInterface {
	return s.runtime
}

type PLUGIN_IN_STREAM_EVENT string

const (
	PLUGIN_IN_STREAM_EVENT_REQUEST  PLUGIN_IN_STREAM_EVENT = "request"
	PLUGIN_IN_STREAM_EVENT_RESPONSE PLUGIN_IN_STREAM_EVENT = "backwards_response"
)

func (s *Session) Message(event PLUGIN_IN_STREAM_EVENT, data any) []byte {
	return parser.MarshalJsonBytes(map[string]any{
		"session_id": s.ID,
		"event":      event,
		"data":       data,
	})
}

func (s *Session) Write(event PLUGIN_IN_STREAM_EVENT, data any) error {
	if s.runtime == nil {
		return errors.New("runtime not bound")
	}
	s.runtime.Write(s.ID, s.Message(event, data))
	return nil
}
