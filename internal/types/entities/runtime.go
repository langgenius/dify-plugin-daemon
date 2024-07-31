package entities

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"encoding/hex"
	"hash/fnv"
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
		// returns the plugin configuration
		Configuration() *plugin_entities.PluginDeclaration
		// unique identity of the plugin
		Identity() (string, error)
		// hashed identity of the plugin
		HashedIdentity() (string, error)
		// before the plugin starts, it will call this method to initialize the environment
		InitEnvironment() error
		// start the plugin, returns errors if the plugin fails to start and hangs until the plugin stops
		StartPlugin() error
		// returns true if the plugin is stopped
		Stopped() bool
		// stop the plugin
		Stop()
		// add a function to be called when the plugin stops
		OnStop(func())
		// trigger the stop event
		TriggerStop()
		// returns the runtime state of the plugin
		RuntimeState() PluginRuntimeState
		// Update the runtime state of the plugin
		UpdateState(state PluginRuntimeState)
		// returns the checksum of the plugin
		Checksum() string
		// wait for the plugin to stop
		Wait() (<-chan bool, error)
		// returns the runtime type of the plugin
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

func HashedIdentity(identity string) string {
	hash := sha256.New()
	hash.Write([]byte(identity))
	return hex.EncodeToString(hash.Sum(nil))
}

func (r *PluginRuntime) HashedIdentity() (string, error) {
	return HashedIdentity(r.Config.Identity()), nil
}

func (r *PluginRuntime) RuntimeState() PluginRuntimeState {
	return r.State
}

func (r *PluginRuntime) UpdateState(state PluginRuntimeState) {
	r.State = state
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
	Identity     string     `json:"identity"`
	Restarts     int        `json:"restarts"`
	Status       string     `json:"status"`
	RelativePath string     `json:"relative_path"`
	ActiveAt     *time.Time `json:"active_at"`
	StoppedAt    *time.Time `json:"stopped_at"`
	Verified     bool       `json:"verified"`
	ScheduledAt  *time.Time `json:"scheduled_at"`
	Logs         []string   `json:"logs"`
}

func (s *PluginRuntimeState) Hash() (uint64, error) {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(s)
	if err != nil {
		return 0, err
	}
	j := fnv.New64a()
	_, err = j.Write(buf.Bytes())
	if err != nil {
		return 0, err
	}

	return j.Sum64(), nil
}

const (
	PLUGIN_RUNTIME_STATUS_ACTIVE     = "active"
	PLUGIN_RUNTIME_STATUS_LAUNCHING  = "launching"
	PLUGIN_RUNTIME_STATUS_STOPPED    = "stopped"
	PLUGIN_RUNTIME_STATUS_RESTARTING = "restarting"
	PLUGIN_RUNTIME_STATUS_PENDING    = "pending"
)
