package entities

import (
	"time"
)

type (
	PluginRuntime struct {
		State     PluginRuntimeState  `json:"state"`
		Config    PluginConfiguration `json:"config"`
		Connector PluginConnector     `json:"-"`
	}

	PluginRuntimeInterface interface {
		PluginRuntimeTimeLifeInterface
		PluginRuntimeSessionIOInterface
	}

	PluginRuntimeTimeLifeInterface interface {
		InitEnvironment() error
		StartPlugin() error
		Stopped() bool
		Stop()
		Configuration() *PluginConfiguration
	}

	PluginRuntimeSessionIOInterface interface {
		Listen(session_id string) *BytesIOListener
		Write(session_id string, data []byte)
		Request(session_id string, data []byte) ([]byte, error)
	}
)

func (r *PluginRuntime) Stopped() bool {
	return r.State.Status == PLUGIN_RUNTIME_STATUS_STOPPED
}

func (r *PluginRuntime) Stop() {
	r.State.Status = PLUGIN_RUNTIME_STATUS_STOPPED
}

func (r *PluginRuntime) Configuration() *PluginConfiguration {
	return &r.Config
}

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

type PluginConnector interface {
	OnMessage(func([]byte))
	Read([]byte) int
	Write([]byte) int
}
