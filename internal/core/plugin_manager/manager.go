package plugin_manager

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

func List() []entities.PluginRuntimeInterface {
	var runtimes []entities.PluginRuntimeInterface
	m.Range(func(key, value interface{}) bool {
		if v, ok := value.(entities.PluginRuntimeInterface); ok {
			runtimes = append(runtimes, v)
		}
		return true
	})
	return runtimes
}

func Get(identity string) entities.PluginRuntimeInterface {
	if v, ok := m.Load(identity); ok {
		if r, ok := v.(entities.PluginRuntimeInterface); ok {
			return r
		}
	}
	return nil
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

	// init redis client
	if err := cache.InitRedisClient(
		fmt.Sprintf("%s:%d", configuration.RedisHost, configuration.RedisPort),
		configuration.RedisPass,
	); err != nil {
		log.Panic("init redis client failed: %s", err.Error())
	}

	if err := dify_invocation.InitDifyInvocationDaemon(
		configuration.PluginInnerApiURL, configuration.PluginInnerApiKey,
	); err != nil {
		log.Panic("init dify invocation daemon failed: %s", err.Error())
	}

	// start plugin watcher
	startWatcher(configuration)

	// start plugin lifetime manager
	startLifeTimeManager(configuration)
}
