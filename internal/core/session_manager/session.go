package session_manager

import (
	"errors"
	"sync"

	"github.com/google/uuid"
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
	id      string
	runtime plugin_entities.PluginRuntimeSessionIOInterface

	tenant_id       string
	user_id         string
	plugin_identity string
	cluster_id      string
}

type SessionInfo struct {
	TenantID       string `json:"tenant_id"`
	UserID         string `json:"user_id"`
	PluginIdentity string `json:"plugin_identity"`
	ClusterID      string `json:"cluster_id"`
}

const (
	SESSION_INFO_MAP_KEY = "session_info"
)

func NewSession(tenant_id string, user_id string, plugin_identity string, cluster_id string) *Session {
	s := &Session{
		id:              uuid.New().String(),
		tenant_id:       tenant_id,
		user_id:         user_id,
		plugin_identity: plugin_identity,
		cluster_id:      cluster_id,
	}

	session_lock.Lock()
	sessions[s.id] = s
	session_lock.Unlock()

	session_info := &SessionInfo{
		TenantID:       tenant_id,
		UserID:         user_id,
		PluginIdentity: plugin_identity,
		ClusterID:      cluster_id,
	}

	if err := cache.SetMapOneField(SESSION_INFO_MAP_KEY, s.id, session_info); err != nil {
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

	if err := cache.DelMapField(SESSION_INFO_MAP_KEY, id); err != nil {
		log.Error("delete session info from cache failed, %s", err)
	}
}

func (s *Session) Close() {
	session_lock.Lock()
	delete(sessions, s.id)
	session_lock.Unlock()

	if err := cache.DelMapField(SESSION_INFO_MAP_KEY, s.id); err != nil {
		log.Error("delete session info from cache failed, %s", err)
	}
}

func (s *Session) ID() string {
	return s.id
}

func (s *Session) TenantID() string {
	return s.tenant_id
}

func (s *Session) UserID() string {
	return s.user_id
}

func (s *Session) PluginIdentity() string {
	return s.plugin_identity
}

func (s *Session) BindRuntime(runtime plugin_entities.PluginRuntimeSessionIOInterface) {
	s.runtime = runtime
}

type PLUGIN_IN_STREAM_EVENT string

const (
	PLUGIN_IN_STREAM_EVENT_REQUEST  PLUGIN_IN_STREAM_EVENT = "request"
	PLUGIN_IN_STREAM_EVENT_RESPONSE PLUGIN_IN_STREAM_EVENT = "backwards_response"
)

func (s *Session) Message(event PLUGIN_IN_STREAM_EVENT, data any) []byte {
	return parser.MarshalJsonBytes(map[string]any{
		"session_id": s.id,
		"event":      event,
		"data":       data,
	})
}

func (s *Session) Write(event PLUGIN_IN_STREAM_EVENT, data any) error {
	if s.runtime == nil {
		return errors.New("runtime not bound")
	}
	s.runtime.Write(s.id, s.Message(event, data))
	return nil
}
