package exception

import "github.com/langgenius/dify-plugin-daemon/pkg/entities"

type PluginDaemonError interface {
	error

	ToResponse() *entities.Response
}
