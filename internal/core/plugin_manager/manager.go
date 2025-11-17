package plugin_manager

import (
	"fmt"
	"strings"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/langgenius/dify-cloud-kit/oss"
	controlpanel "github.com/langgenius/dify-plugin-daemon/internal/core/control_panel"
	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation/calldify"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/media_transport"
	serverless "github.com/langgenius/dify-plugin-daemon/internal/core/serverless_connector"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

type PluginManager struct {
	// mediaBucket is used to manage media files like plugin icons, images, etc.
	mediaBucket *media_transport.MediaBucket

	// packageBucket can be considered as a collection of all uploaded plugin
	// original packages, once a package was accepted by Dify,
	// it should be stored here.
	packageBucket *media_transport.PackageBucket

	// installedBucket is used to manage installed plugins, all the installed plugins will be saved here
	// `accepted` dose not means `installed`, a installed plugin
	// will be scheduled by daemon, daemon copied and move the package
	// from `packageBucket` to `installedBucket`
	// the copy processing marks a plugin as `installed`
	// as for `scheduling`, it's automatically done by control panel
	// of course you may use `controlPanel.LaunchLocalPlugin` to start it manually
	installedBucket *media_transport.InstalledBucket

	// backwardsInvocation is a handle to invoke dify
	backwardsInvocation dify_invocation.BackwardsInvocation

	config *app.Config

	// plugin lifecycle controller
	//
	// whatever it's local mode or serverless mode, all the signals and calls
	// which related to plugin lifecycle should be handled by it.
	// so that we can decouple lifetime control and thirdparty service like package management
	controlPanel *controlpanel.ControlPanel
}

var (
	manager *PluginManager
)

func InitGlobalManager(oss oss.OSS, config *app.Config) *PluginManager {
	mediaBucket := media_transport.NewAssetsBucket(
		oss,
		config.PluginMediaCachePath,
		config.PluginMediaCacheSize,
	)

	installedBucket := media_transport.NewInstalledBucket(
		oss,
		config.PluginInstalledPath,
	)

	packageBucket := media_transport.NewPackageBucket(
		oss,
		config.PluginPackageCachePath,
	)

	manager = &PluginManager{
		mediaBucket:     mediaBucket,
		packageBucket:   packageBucket,
		installedBucket: installedBucket,
		controlPanel: controlpanel.NewControlPanel(
			config,
			mediaBucket,
			packageBucket,
			installedBucket,
		),
		config: config,
	}

	// mount logger to control panel
	manager.controlPanel.AddNotifier(&controlpanel.StandardLogger{})

	return manager
}

func Manager() *PluginManager {
	return manager
}

func (p *PluginManager) GetAsset(id string) ([]byte, error) {
	return p.mediaBucket.Get(id)
}

func (p *PluginManager) Launch(configuration *app.Config) {
	log.Info("start plugin manager daemon...")

	// init redis client
	if configuration.RedisUseSentinel {
		// use Redis Sentinel
		sentinels := strings.Split(configuration.RedisSentinels, ",")
		if err := cache.InitRedisSentinelClient(
			sentinels,
			configuration.RedisSentinelServiceName,
			configuration.RedisUser,
			configuration.RedisPass,
			configuration.RedisSentinelUsername,
			configuration.RedisSentinelPassword,
			configuration.RedisUseSsl,
			configuration.RedisDB,
			configuration.RedisSentinelSocketTimeout,
		); err != nil {
			log.Panic("init redis sentinel client failed: %s", err.Error())
		}
	} else {
		if err := cache.InitRedisClient(
			fmt.Sprintf("%s:%d", configuration.RedisHost, configuration.RedisPort),
			configuration.RedisUser,
			configuration.RedisPass,
			configuration.RedisUseSsl,
			configuration.RedisDB,
		); err != nil {
			log.Panic("init redis client failed: %s", err.Error())
		}
	}

	invocation, err := calldify.NewDifyInvocationDaemon(
		calldify.NewDifyInvocationDaemonPayload{
			BaseUrl:      configuration.DifyInnerApiURL,
			CallingKey:   configuration.DifyInnerApiKey,
			WriteTimeout: configuration.DifyInvocationWriteTimeout,
			ReadTimeout:  configuration.DifyInvocationReadTimeout,
		},
	)
	if err != nil {
		log.Panic("init dify invocation daemon failed: %s", err.Error())
	}
	p.backwardsInvocation = invocation

	// start control panel
	p.controlPanel.StartWatchDog()

	// launch serverless connector
	if configuration.Platform == app.PLATFORM_SERVERLESS {
		serverless.Init(configuration)
	}
}

func (p *PluginManager) BackwardsInvocation() dify_invocation.BackwardsInvocation {
	return p.backwardsInvocation
}

// check if the plugin is already running on this node
func (c *PluginManager) NeedRedirecting(
	identity plugin_entities.PluginUniqueIdentifier,
) (bool, error) {
	// debugging runtime were stored in control panel
	if identity.RemoteLike() {
		_, err := c.controlPanel.GetPluginRuntime(identity)
		if err != nil {
			return true, err
		}
		return false, nil
	}

	if c.config.Platform == app.PLATFORM_SERVERLESS {
		// under serverless mode, it's no need to do redirecting
		return false, nil
	} else if c.config.Platform == app.PLATFORM_LOCAL {
		// under local mode, check if the plugin is already running on this node
		_, err := c.controlPanel.GetPluginRuntime(identity)
		if err != nil {
			// not found on this node, need to redirecting
			return true, err
		}

		// found on current node
		return false, nil
	}

	return true, nil
}

var (
	pluginAssetCache *lru.Cache[string, []byte]
)

func pluginAssetCacheKey(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	path string,
) string {
	return fmt.Sprintf("%s/%s", pluginUniqueIdentifier.String(), path)
}
func (p *PluginManager) ExtractPluginAsset(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	path string,
) ([]byte, error) {
	key := pluginAssetCacheKey(pluginUniqueIdentifier, path)
	cached, ok := pluginAssetCache.Get(key)
	if ok {
		return cached, nil
	}
	pkgBytes, err := manager.GetPackage(pluginUniqueIdentifier)
	if err != nil {
		return nil, err
	}
	zipDecoder, err := decoder.NewZipPluginDecoder(pkgBytes)
	if err != nil {
		return nil, err
	}
	assets, err := zipDecoder.Assets()
	if err != nil {
		return nil, err
	}
	pluginAssetCache.Add(key, assets[path])
	return assets[path], nil
}
