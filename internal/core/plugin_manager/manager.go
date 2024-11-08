package plugin_manager

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation"
	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation/real"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/media_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/remote_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/serverless"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/oss"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache/helper"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/lock"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/mapping"
)

type PluginManager struct {
	m mapping.Map[string, plugin_entities.PluginLifetime]

	// max size of a plugin package
	maxPluginPackageSize int64

	// where the plugin finally running
	workingDirectory string

	// where the plugin finally installed but not running
	pluginStoragePath string

	// mediaBucket is used to manage media files like plugin icons, images, etc.
	mediaBucket *media_manager.MediaBucket

	// packageBucket is used to manage plugin packages, all the packages uploaded by users will be saved here
	packageBucket *media_manager.PackageBucket

	// installedBucket is used to manage installed plugins, all the installed plugins will be saved here
	installedBucket *media_manager.InstalledBucket

	// register plugin
	pluginRegisters []func(lifetime plugin_entities.PluginLifetime) error

	// localPluginLaunchingLock is a lock to launch local plugins
	localPluginLaunchingLock *lock.GranularityLock

	// backwardsInvocation is a handle to invoke dify
	backwardsInvocation dify_invocation.BackwardsInvocation

	// python interpreter path
	pythonInterpreterPath string

	// remote plugin server
	remotePluginServer remote_manager.RemotePluginServerInterface

	// max launching lock to prevent too many plugins launching at the same time
	maxLaunchingLock chan bool
}

var (
	manager *PluginManager
)

func InitGlobalManager(oss oss.OSS, configuration *app.Config) *PluginManager {
	manager = &PluginManager{
		maxPluginPackageSize: configuration.MaxPluginPackageSize,
		pluginStoragePath:    configuration.PluginInstalledPath,
		workingDirectory:     configuration.PluginWorkingPath,
		mediaBucket: media_manager.NewAssetsBucket(
			oss,
			configuration.PluginMediaCachePath,
			configuration.PluginMediaCacheSize,
		),
		packageBucket: media_manager.NewPackageBucket(
			oss,
			configuration.PluginPackageCachePath,
		),
		installedBucket: media_manager.NewInstalledBucket(
			oss,
			configuration.PluginInstalledPath,
		),
		localPluginLaunchingLock: lock.NewGranularityLock(),
		maxLaunchingLock:         make(chan bool, 2), // by default, we allow 2 plugins launching at the same time
		pythonInterpreterPath:    configuration.PythonInterpreterPath,
	}

	// mkdir
	os.MkdirAll(configuration.PluginWorkingPath, 0755)
	os.MkdirAll(configuration.PluginInstalledPath, 0755)
	os.MkdirAll(configuration.PluginMediaCachePath, 0755)
	os.MkdirAll(configuration.PluginPackageCachePath, 0755)
	os.MkdirAll(filepath.Dir(configuration.ProcessCachingPath), 0755)

	return manager
}

func Manager() *PluginManager {
	return manager
}

func (p *PluginManager) Get(
	identity plugin_entities.PluginUniqueIdentifier,
) (plugin_entities.PluginLifetime, error) {
	if v, ok := p.m.Load(identity.String()); ok {
		return v, nil
	}

	// check if plugin is a serverless runtime
	plugin_session_interface, err := p.getServerlessPluginRuntime(identity)
	if err != nil {
		return nil, err
	}

	return plugin_session_interface, nil
}

func (p *PluginManager) GetAsset(id string) ([]byte, error) {
	return p.mediaBucket.Get(id)
}

func (p *PluginManager) Launch(configuration *app.Config) {
	log.Info("start plugin manager daemon...")

	// init redis client
	if err := cache.InitRedisClient(
		fmt.Sprintf("%s:%d", configuration.RedisHost, configuration.RedisPort),
		configuration.RedisPass,
	); err != nil {
		log.Panic("init redis client failed: %s", err.Error())
	}

	invocation, err := real.NewDifyInvocationDaemon(
		configuration.DifyInnerApiURL, configuration.DifyInnerApiKey,
	)
	if err != nil {
		log.Panic("init dify invocation daemon failed: %s", err.Error())
	}
	p.backwardsInvocation = invocation

	// start local watcher
	if configuration.Platform == app.PLATFORM_LOCAL {
		p.startLocalWatcher()
	}

	// launch serverless connector
	if configuration.Platform == app.PLATFORM_AWS_LAMBDA {
		serverless.Init(configuration)
	}

	// start remote watcher
	p.startRemoteWatcher(configuration)
}

func (p *PluginManager) BackwardsInvocation() dify_invocation.BackwardsInvocation {
	return p.backwardsInvocation
}

func (p *PluginManager) SavePackage(plugin_unique_identifier plugin_entities.PluginUniqueIdentifier, pkg []byte) (
	*plugin_entities.PluginDeclaration, error,
) {
	// save to storage
	err := p.packageBucket.Save(plugin_unique_identifier.String(), pkg)
	if err != nil {
		return nil, err
	}

	// try to decode the package
	package_decoder, err := decoder.NewZipPluginDecoder(pkg)
	if err != nil {
		return nil, err
	}

	// get the declaration
	declaration, err := package_decoder.Manifest()
	if err != nil {
		return nil, err
	}

	// get the assets
	assets, err := package_decoder.Assets()
	if err != nil {
		return nil, err
	}

	// remap the assets
	_, err = p.mediaBucket.RemapAssets(&declaration, assets)
	if err != nil {
		return nil, err
	}

	unique_identifier, err := package_decoder.UniqueIdentity()
	if err != nil {
		return nil, err
	}

	// create plugin if not exists
	if _, err := db.GetOne[models.PluginDeclaration](
		db.Equal("plugin_unique_identifier", unique_identifier.String()),
	); err == db.ErrDatabaseNotFound {
		err = db.Create(&models.PluginDeclaration{
			PluginUniqueIdentifier: unique_identifier.String(),
			PluginID:               unique_identifier.PluginID(),
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
) (
	*plugin_entities.PluginDeclaration, error,
) {
	return helper.CombinedGetPluginDeclaration(plugin_unique_identifier)
}
