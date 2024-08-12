package entities

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"hash/fnv"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
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
		PluginRuntimeDockerInterface
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
		UpdateScheduledAt(t time.Time)
		// returns the checksum of the plugin
		Checksum() string
		// wait for the plugin to stop
		Wait() (<-chan bool, error)
		// returns the runtime type of the plugin
		Type() PluginRuntimeType

		// set the plugin to active
		SetActive()
		// set the plugin to launching
		SetLaunching()
		// set the plugin to restarting
		SetRestarting()
		// set the plugin to pending
		SetPending()
		// set the active time of the plugin
		SetActiveAt(t time.Time)
		// set the scheduled time of the plugin
		SetScheduledAt(t time.Time)
		// add restarts to the plugin
		AddRestarts()
	}

	PluginRuntimeSessionIOInterface interface {
		Listen(session_id string) *BytesIOListener
		Write(session_id string, data []byte)
	}

	PluginRuntimeDockerInterface interface {
		// returns the docker image and the delete function, always call the delete function when the image is no longer needed
		DockerImage() (string, func(), error)
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

func (r *PluginRuntime) UpdateScheduledAt(t time.Time) {
	r.State.ScheduledAt = &t
}

func (r *PluginRuntime) InitState() {
	r.State = PluginRuntimeState{
		Restarts:    0,
		Status:      PLUGIN_RUNTIME_STATUS_PENDING,
		ActiveAt:    nil,
		StoppedAt:   nil,
		Verified:    false,
		ScheduledAt: nil,
		Logs:        []string{},
	}
}

func (r *PluginRuntime) SetActive() {
	r.State.Status = PLUGIN_RUNTIME_STATUS_ACTIVE
}

func (r *PluginRuntime) SetLaunching() {
	r.State.Status = PLUGIN_RUNTIME_STATUS_LAUNCHING
}

func (r *PluginRuntime) SetRestarting() {
	r.State.Status = PLUGIN_RUNTIME_STATUS_RESTARTING
}

func (r *PluginRuntime) SetPending() {
	r.State.Status = PLUGIN_RUNTIME_STATUS_PENDING
}

func (r *PluginRuntime) SetActiveAt(t time.Time) {
	r.State.ActiveAt = &t
}

func (r *PluginRuntime) SetScheduledAt(t time.Time) {
	r.State.ScheduledAt = &t
}

func (r *PluginRuntime) AddRestarts() {
	r.State.Restarts++
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
	AbsolutePath string     `json:"absolute_path"`
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
