package app

import "golang.org/x/exp/constraints"

func (config *Config) SetDefault() {
	setDefaultInt(&config.ServerPort, 5002)
	setDefaultInt(&config.RoutinePoolSize, 10000)
	setDefaultInt(&config.LifetimeCollectionGCInterval, 60)
	setDefaultInt(&config.LifetimeCollectionHeartbeatInterval, 5)
	setDefaultInt(&config.LifetimeStateGCInterval, 300)
	setDefaultInt(&config.DifyInvocationConnectionIdleTimeout, 120)
	setDefaultInt(&config.PluginRemoteInstallServerEventLoopNums, 8)
	setDefaultInt(&config.PluginRemoteInstallingMaxConn, 256)
	setDefaultInt(&config.MaxPluginPackageSize, 52428800)
	setDefaultInt(&config.MaxBundlePackageSize, 52428800*12)
	setDefaultInt(&config.MaxServerlessTransactionTimeout, 300)
	setDefaultInt(&config.PluginMaxExecutionTimeout, 10*60)
	setDefaultString(&config.PluginStorageType, "local")
	setDefaultInt(&config.PluginMediaCacheSize, 1024)
	setDefaultInt(&config.PluginRemoteInstallingMaxSingleTenantConn, 5)
	setDefaultBoolPtr(&config.PluginRemoteInstallingEnabled, true)
	setDefaultBoolPtr(&config.PluginEndpointEnabled, true)
	setDefaultString(&config.DBSslMode, "disable")
	setDefaultString(&config.PluginStorageLocalRoot, "storage")
	setDefaultString(&config.PluginInstalledPath, "plugin")
	setDefaultString(&config.PluginMediaCachePath, "assets")
	setDefaultString(&config.PersistenceStoragePath, "persistence")
	setDefaultInt(&config.PluginLocalLaunchingConcurrent, 2)
	setDefaultInt(&config.PersistenceStorageMaxSize, 100*1024*1024)
	setDefaultString(&config.PluginPackageCachePath, "plugin_packages")
	setDefaultString(&config.PythonInterpreterPath, "/usr/bin/python3")
	setDefaultInt(&config.PythonEnvInitTimeout, 120)
	setDefaultBoolPtr(&config.ForceVerifyingSignature, true)
	setDefaultBoolPtr(&config.PipPreferBinary, true)
	setDefaultBoolPtr(&config.PipVerbose, true)
	if config.DBType == "postgresql" {
		setDefaultString(&config.DBDefaultDatabase, "postgres")
	} else if config.DBType == "mysql" {
		setDefaultString(&config.DBDefaultDatabase, "mysql")
	}
	setDefaultBoolPtr(&config.HealthApiLogEnabled, true)
}

func setDefaultInt[T constraints.Integer](value *T, defaultValue T) {
	if *value == 0 {
		*value = defaultValue
	}
}

func setDefaultString(value *string, defaultValue string) {
	if *value == "" {
		*value = defaultValue
	}
}

func setDefaultBoolPtr(value **bool, defaultValue bool) {
	if *value == nil {
		*value = &defaultValue
	}
}
