package session_manager

import (
	"sync"

	"github.com/google/uuid"
)

var (
	sessions     map[string]*Session = map[string]*Session{}
	session_lock sync.RWMutex
)

type Session struct {
	id string

	tenant_id       string
	user_id         string
	plugin_identity string
}

func NewSession(tenant_id string, user_id string, plugin_identity string) *Session {
	s := &Session{
		id:              uuid.New().String(),
		tenant_id:       tenant_id,
		user_id:         user_id,
		plugin_identity: plugin_identity,
	}

	session_lock.Lock()
	defer session_lock.Unlock()

	sessions[s.id] = s

	return s
}

func GetSession(id string) *Session {
	session_lock.RLock()
	defer session_lock.RUnlock()

	return sessions[id]
}

func DeleteSession(id string) {
	session_lock.Lock()
	defer session_lock.Unlock()

	delete(sessions, id)
}

func (s *Session) Close() {
	session_lock.Lock()
	defer session_lock.Unlock()

	delete(sessions, s.id)
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
