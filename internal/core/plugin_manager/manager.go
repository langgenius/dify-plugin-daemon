package plugin_manager

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/langgenius/dify-cloud-kit/oss"
	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation/real"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/debugging_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/media_transport"
	serverless "github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/serverless_connector"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache/helper"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/lock"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/mapping"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

type PluginManager struct {
	m mapping.Map[string, plugin_entities.PluginLifetime]

	// mediaBucket is used to manage media files like plugin icons, images, etc.
	mediaBucket *media_transport.MediaBucket

	// packageBucket is used to manage plugin packages, all the packages uploaded by users will be saved here
	packageBucket *media_transport.PackageBucket

	// installedBucket is used to manage installed plugins, all the installed plugins will be saved here
	installedBucket *media_transport.InstalledBucket

	// register plugin
	pluginRegisters []func(lifetime plugin_entities.PluginLifetime) error

	// localPluginLaunchingLock is a lock to launch local plugins
	localPluginLaunchingLock *lock.GranularityLock

	// backwardsInvocation is a handle to invoke dify
	backwardsInvocation dify_invocation.BackwardsInvocation

	config *app.Config

	// remote plugin server
	remotePluginServer debugging_runtime.RemotePluginServerInterface

	// max launching lock to prevent too many plugins launching at the same time
	maxLaunchingLock chan bool
}

var (
	manager *PluginManager
)

func InitGlobalManager(oss oss.OSS, configuration *app.Config) *PluginManager {
	manager = &PluginManager{
		mediaBucket: media_transport.NewAssetsBucket(
			oss,
			configuration.PluginMediaCachePath,
			configuration.PluginMediaCacheSize,
		),
		packageBucket: media_transport.NewPackageBucket(
			oss,
			configuration.PluginPackageCachePath,
		),
		installedBucket: media_transport.NewInstalledBucket(
			oss,
			configuration.PluginInstalledPath,
		),
		localPluginLaunchingLock: lock.NewGranularityLock(),
		// By default, we allow up to configuration.PluginLocalLaunchingConcurrent plugins to be launched concurrently; if not configured, the default is 2.
		maxLaunchingLock: make(chan bool, configuration.PluginLocalLaunchingConcurrent),
		config:           configuration,
	}

	return manager
}

func Manager() *PluginManager {
	return manager
}

func (p *PluginManager) Get(
	identity plugin_entities.PluginUniqueIdentifier,
) (plugin_entities.PluginLifetime, error) {
	if identity.RemoteLike() || p.config.Platform == app.PLATFORM_LOCAL {
		// check if it's a debugging plugin or a local plugin
		if v, ok := p.m.Load(identity.String()); ok {
			return v, nil
		}
		return nil, errors.New("plugin not found")
	} else {
		// otherwise, use serverless runtime instead
		pluginSessionInterface, err := p.getServerlessPluginRuntime(identity)
		if err != nil {
			return nil, err
		}

		return pluginSessionInterface, nil
	}
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

	invocation, err := real.NewDifyInvocationDaemon(
		real.NewDifyInvocationDaemonPayload{
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

	// start local watcher
	if configuration.Platform == app.PLATFORM_LOCAL {
		p.startLocalWatcher(configuration)
	}

	// launch serverless connector
	if configuration.Platform == app.PLATFORM_SERVERLESS {
		serverless.Init(configuration)
	}

	// start remote watcher
	p.startRemoteWatcher(configuration)
}

func (p *PluginManager) BackwardsInvocation() dify_invocation.BackwardsInvocation {
	return p.backwardsInvocation
}

func (p *PluginManager) SavePackage(plugin_unique_identifier plugin_entities.PluginUniqueIdentifier, pkg []byte, thirdPartySignatureVerificationConfig *decoder.ThirdPartySignatureVerificationConfig) (
	*plugin_entities.PluginDeclaration, error,
) {
	// try to decode the package
	packageDecoder, err := decoder.NewZipPluginDecoderWithThirdPartySignatureVerificationConfig(pkg, thirdPartySignatureVerificationConfig)
	if err != nil {
		return nil, err
	}

	// get the declaration
	declaration, err := packageDecoder.Manifest()
	if err != nil {
		return nil, err
	}

	if err := declaration.ManifestValidate(); err != nil {
		return nil, errors.Join(err, fmt.Errorf("illegal plugin manifest"))
	}

	// get the assets
	assets, err := packageDecoder.Assets()
	if err != nil {
		return nil, err
	}

	// remap the assets
	_, err = p.mediaBucket.RemapAssets(&declaration, assets)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("failed to remap assets"))
	}

	uniqueIdentifier, err := packageDecoder.UniqueIdentity()
	if err != nil {
		return nil, err
	}

	// save to storage
	err = p.packageBucket.Save(plugin_unique_identifier.String(), pkg)
	if err != nil {
		return nil, err
	}

	// create plugin if not exists
	if _, err := db.GetOne[models.PluginDeclaration](
		db.Equal("plugin_unique_identifier", uniqueIdentifier.String()),
	); err == db.ErrDatabaseNotFound {
		err = db.Create(&models.PluginDeclaration{
			PluginUniqueIdentifier: uniqueIdentifier.String(),
			PluginID:               uniqueIdentifier.PluginID(),
			Declaration:            declaration,
		})
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return &declaration, nil
}

func (p *PluginManager) GetPackage(
	plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
) ([]byte, error) {
	file, err := p.packageBucket.Get(plugin_unique_identifier.String())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("plugin package not found, please upload it firstly")
		}
		return nil, err
	}

	return file, nil
}

func (p *PluginManager) GetDeclaration(
	plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
	tenant_id string,
	runtime_type plugin_entities.PluginRuntimeType,
) (
	*plugin_entities.PluginDeclaration, error,
) {
	return helper.CombinedGetPluginDeclaration(
		plugin_unique_identifier, runtime_type,
	)
}
