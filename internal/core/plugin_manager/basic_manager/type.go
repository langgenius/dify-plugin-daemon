package basic_manager

import "github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/media_manager"

type BasicPluginRuntime struct {
	mediaManager *media_manager.MediaBucket

	assetsIds []string
}

func NewBasicPluginRuntime(mediaManager *media_manager.MediaBucket) BasicPluginRuntime {
	return BasicPluginRuntime{mediaManager: mediaManager}
}
