package exception

import "github.com/langgenius/dify-plugin-daemon/internal/types/entities"

type PluginDaemonError interface {
	error

	ToResponse() *entities.Response
}
