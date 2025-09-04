package curd

import "github.com/langgenius/dify-plugin-daemon/internal/types/app"

var (
	allowOrphans bool
)

func Init(config *app.Config) {
	allowOrphans = config.PluginAllowOrphans
}
