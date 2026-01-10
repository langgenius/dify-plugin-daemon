package app

import (
	"os"

	"github.com/langgenius/dify-cloud-kit/oss"
	"golang.org/x/exp/constraints"
)

func (config *Config) SetDefault() {
	switch config.DBType {
	case DB_TYPE_OCEANBASE, DB_TYPE_SEEKDB:
		config.DBType = DB_TYPE_MYSQL
	}
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
	setDefaultString(&config.PluginStorageType, oss.OSS_TYPE_LOCAL)
	setDefaultInt(&config.PluginMediaCacheSize, 1024)
	setDefaultInt(&config.DifyPluginServerlessConnectorLaunchTimeout, 240)
	setDefaultInt(&config.PluginRemoteInstallingMaxSingleTenantConn, 5)
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
	setDefaultInt(&config.DifyInvocationWriteTimeout, 5000)
	setDefaultInt(&config.DifyInvocationReadTimeout, 240000)

	// fallback to lowercase proxy environment variables if uppercase is not set
	if _, ok := os.LookupEnv("HTTP_PROXY"); !ok {
		config.HttpProxy = os.Getenv("http_proxy")
	}
	if _, ok := os.LookupEnv("HTTPS_PROXY"); !ok {
		config.HttpsProxy = os.Getenv("https_proxy")
	}
	if _, ok := os.LookupEnv("NO_PROXY"); !ok {
		config.NoProxy = os.Getenv("no_proxy")
	}
	if config.DBType == DB_TYPE_POSTGRESQL || config.DBType == DB_TYPE_PG_BOUNCER {
		setDefaultString(&config.DBDefaultDatabase, "postgres")
	} else if config.DBType == DB_TYPE_MYSQL {
		setDefaultString(&config.DBDefaultDatabase, "mysql")
	}
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
