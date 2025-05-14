package run

import (
	"context"
	"io"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/access_types"
)

type RunMode = string

const (
	RUN_MODE_STDIO RunMode = "stdio"
	RUN_MODE_TCP   RunMode = "tcp"
)

type RunPluginPayload struct {
	PluginPath string
	RunMode    RunMode
	EnableLogs bool

	TcpServerPort int
	TcpServerHost string
}

type client struct {
	id     string
	reader io.ReadCloser
	writer io.WriteCloser
	cancel context.CancelFunc

	onClose func()
}

func (c *client) OnClose(fn func()) {
	c.onClose = fn
}

type InvokePluginPayload struct {
	Type   access_types.PluginAccessType   `json:"type"`
	Action access_types.PluginAccessAction `json:"action"`

	Request map[string]any `json:"request"`
}

type GenericResponseType = string

const (
	GENERIC_RESPONSE_TYPE_INFO            GenericResponseType = "info"
	GENERIC_RESPONSE_TYPE_ERROR           GenericResponseType = "error"
	GENERIC_RESPONSE_TYPE_PLUGIN_RESPONSE GenericResponseType = "plugin_response"
)

type GenericResponse struct {
	Type GenericResponseType `json:"type"`

	Response map[string]any `json:"response"`
}
