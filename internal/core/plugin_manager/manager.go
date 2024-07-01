package plugin_manager

import (
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

func List() []*entities.PluginRuntime {
	var runtimes []*entities.PluginRuntime
	m.Range(func(key, value interface{}) bool {
		if v, ok := value.(*entities.PluginRuntime); ok {
			runtimes = append(runtimes, v)
		}
		return true
	})
	return runtimes
}

func Put(path string, binary []byte) {
	//TODO: put binary into
}

func Delete(identity string) {
	//TODO: delete binary from
}

func Init(configuration *app.Config) {
	// TODO: init plugin manager
	log.Info("start plugin manager daemon...")

	startWatcher(configuration.StoragePath)
}
