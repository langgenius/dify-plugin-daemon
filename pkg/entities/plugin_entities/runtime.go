package plugin_entities

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/access_types"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
)

type (
	PluginRuntime struct {
		State     PluginRuntimeState `json:"state"`
		Config    PluginDeclaration  `json:"config"`
		onStopped []func()           `json:"-"`
	}

	PluginLifetime interface {
		PluginBasicInfoInterface
		PluginRuntimeSessionIOInterface
		PluginClusterLifetime
	}

	PluginFullDuplexLifetime interface {
		PluginLifetime

		// before the plugin starts, it will call this method to initialize the environment
		InitEnvironment() error
		// start the plugin, returns errors if the plugin fails to start and hangs until the plugin stops
		StartPlugin() error
		// wait for the plugin to stop
		Wait() (<-chan bool, error)
		// Cleanup the plugin runtime
		Cleanup()
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
		// Started
		WaitStarted() <-chan bool
		// Stopped
		WaitStopped() <-chan bool
	}

	PluginServerlessLifetime interface {
		PluginLifetime

		// before the plugin starts, it will call this method to initialize the environment
		InitEnvironment() error
		// UploadPlugin uploads the plugin to the AWS Lambda
		UploadPlugin() error
	}

	PluginRuntimeSessionIOInterface interface {
		PluginBasicInfoInterface

		// Listen listens for messages from the plugin
		Listen(session_id string) *entities.Broadcast[SessionMessage]
		// Write writes a message to the plugin
		Write(session_id string, action access_types.PluginAccessAction, data []byte)
		// Log adds a log to the plugin runtime state
		Log(string)
		// Warn adds a warning to the plugin runtime state
		Warn(string)
		// Error adds an error to the plugin runtime state
		Error(string)
	}

	PluginClusterLifetime interface {
		// stop the plugin
		Stop()
		// add a function to be called when the plugin stops
		OnStop(func())
		// trigger the stop event
		TriggerStop()
		// returns true if the plugin is stopped
		Stopped() bool
		// returns the runtime state of the plugin
		RuntimeState() PluginRuntimeState
		// Update the runtime state of the plugin
		UpdateScheduledAt(t time.Time)
	}

	PluginBasicInfoInterface interface {
		// returns the runtime type of the plugin
		Type() PluginRuntimeType
		// returns the plugin configuration
		Configuration() *PluginDeclaration
		// unique identity of the plugin
		Identity() (PluginUniqueIdentifier, error)
		// hashed identity of the plugin
		HashedIdentity() (string, error)
		// returns the checksum of the plugin
		Checksum() (string, error)
	}
)

func (r *PluginRuntime) Stopped() bool {
	return r.State.Status == PLUGIN_RUNTIME_STATUS_STOPPED
}

func (r *PluginRuntime) Stop() {
	r.State.Status = PLUGIN_RUNTIME_STATUS_STOPPED
}

func (r *PluginRuntime) Configuration() *PluginDeclaration {
	return &r.Config
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

func (s *PluginRuntime) Log(log string) {
	s.State.Logs = append(s.State.Logs, fmt.Sprintf("[Info] %s: %s", time.Now().Format(time.RFC3339), log))
}

func (s *PluginRuntime) Warn(log string) {
	s.State.Logs = append(s.State.Logs, fmt.Sprintf("[Warn] %s: %s", time.Now().Format(time.RFC3339), log))
}

func (s *PluginRuntime) Error(log string) {
	s.State.Logs = append(s.State.Logs, fmt.Sprintf("[Error] %s: %s", time.Now().Format(time.RFC3339), log))
}

type PluginRuntimeType string

const (
	PLUGIN_RUNTIME_TYPE_LOCAL      PluginRuntimeType = "local"
	PLUGIN_RUNTIME_TYPE_REMOTE     PluginRuntimeType = "remote"
	PLUGIN_RUNTIME_TYPE_SERVERLESS PluginRuntimeType = "serverless"
)

type PluginRuntimeState struct {
	Restarts    int        `json:"restarts"`
	Status      string     `json:"status"`
	WorkingPath string     `json:"working_path"`
	ActiveAt    *time.Time `json:"active_at"`
	StoppedAt   *time.Time `json:"stopped_at"`
	Verified    bool       `json:"verified"`
	ScheduledAt *time.Time `json:"scheduled_at"`
	Logs        []string   `json:"logs"`
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
