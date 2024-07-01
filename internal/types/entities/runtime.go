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
		InitEnvironment() error
		StartPlugin() error
		Stopped() bool
		Stop()
		Configuration() *PluginConfiguration
	}
)

func (r *PluginRuntime) Stopped() bool {
	return r.State.Stopped
}

func (r *PluginRuntime) Stop() {
	r.State.Stopped = true
}

func (r *PluginRuntime) Configuration() *PluginConfiguration {
	return &r.Config
}

type PluginRuntimeState struct {
	Restarts     int        `json:"restarts"`
	Active       bool       `json:"active"`
	RelativePath string     `json:"relative_path"`
	ActiveAt     *time.Time `json:"active_at"`
	DeadAt       *time.Time `json:"dead_at"`
	Stopped      bool       `json:"stopped"`
	Verified     bool       `json:"verified"`
}

type PluginConnector interface {
	OnMessage(func([]byte))
	Read([]byte) int
	Write([]byte) int
}
