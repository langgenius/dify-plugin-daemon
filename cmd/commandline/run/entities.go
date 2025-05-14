package run

import (
	"context"
	"io"
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

type protocol struct {
}
