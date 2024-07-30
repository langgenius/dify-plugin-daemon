package entities

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

type (
	PluginRuntime struct {
		State     PluginRuntimeState                `json:"state"`
		Config    plugin_entities.PluginDeclaration `json:"config"`
		onStopped []func()                          `json:"-"`
	}

	PluginRuntimeInterface interface {
		PluginRuntimeTimeLifeInterface
		PluginRuntimeSessionIOInterface
	}

	PluginRuntimeTimeLifeInterface interface {
		Configuration() *plugin_entities.PluginDeclaration
		Identity() (string, error)
		InitEnvironment() error
		StartPlugin() error
		Stopped() bool
		Stop()
		OnStop(func())
		TriggerStop()
		RuntimeState() *PluginRuntimeState
		Checksum() string
		Wait() (<-chan bool, error)
		Type() PluginRuntimeType
	}

	PluginRuntimeSessionIOInterface interface {
		Listen(session_id string) *BytesIOListener
		Write(session_id string, data []byte)
	}
)

func (r *PluginRuntime) Stopped() bool {
	return r.State.Status == PLUGIN_RUNTIME_STATUS_STOPPED
}

func (r *PluginRuntime) Stop() {
	r.State.Status = PLUGIN_RUNTIME_STATUS_STOPPED
}

func (r *PluginRuntime) Configuration() *plugin_entities.PluginDeclaration {
	return &r.Config
}

func (r *PluginRuntime) Identity() (string, error) {
	return r.Config.Identity(), nil
}

func (r *PluginRuntime) RuntimeState() *PluginRuntimeState {
	return &r.State
}

func (r *PluginRuntime) Checksum() string {
	buf := bytes.Buffer{}
	binary.Write(&buf, binary.BigEndian, parser.MarshalJsonBytes(r.Config))
	hash := sha256.New()
	hash.Write(buf.Bytes())
	return hex.EncodeToString(hash.Sum(nil))
}

func (r *PluginRuntime) OnStop(f func()) {
	r.onStopped = append(r.onStopped, f)
}

func (r *PluginRuntime) TriggerStop() {
	for _, f := range r.onStopped {
		f()
	}
}

type PluginRuntimeType string

const (
	PLUGIN_RUNTIME_TYPE_LOCAL  PluginRuntimeType = "local"
	PLUGIN_RUNTIME_TYPE_REMOTE PluginRuntimeType = "remote"
	PLUGIN_RUNTIME_TYPE_AWS    PluginRuntimeType = "aws"
)

type PluginRuntimeState struct {
	Restarts     int        `json:"restarts"`
	Status       string     `json:"status"`
	RelativePath string     `json:"relative_path"`
	ActiveAt     *time.Time `json:"active_at"`
	StoppedAt    *time.Time `json:"stopped_at"`
	Verified     bool       `json:"verified"`
}

const (
	PLUGIN_RUNTIME_STATUS_ACTIVE     = "active"
	PLUGIN_RUNTIME_STATUS_LAUNCHING  = "launching"
	PLUGIN_RUNTIME_STATUS_STOPPED    = "stopped"
	PLUGIN_RUNTIME_STATUS_RESTARTING = "restarting"
	PLUGIN_RUNTIME_STATUS_PENDING    = "pending"
)
