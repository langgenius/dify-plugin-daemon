package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ClusterDisabled_Default(t *testing.T) {
	config := &Config{}

	// Test default value
	assert.False(t, config.ClusterDisabled)
}

func TestConfig_ClusterDisabled_SetTrue(t *testing.T) {
	config := &Config{
		ClusterDisabled: true,
	}

	assert.True(t, config.ClusterDisabled)
}

func TestConfig_ClusterDisabled_SetFalse(t *testing.T) {
	config := &Config{
		ClusterDisabled: false,
	}

	assert.False(t, config.ClusterDisabled)
}

func TestConfig_Validate_WithClusterDisabled(t *testing.T) {
	config := &Config{
		ServerPort:                          5002,
		ServerKey:                           "test-key",
		DifyInnerApiURL:                     "http://localhost:8000",
		DifyInnerApiKey:                     "test-api-key",
		PluginStorageType:                   "local",
		PluginInstalledPath:                 "/tmp/plugins",
		PluginPackageCachePath:              "/tmp/cache",
		PluginWorkingPath:                   "/tmp/work",
		PluginMaxExecutionTimeout:           300,
		PluginLocalLaunchingConcurrent:      5,
		Platform:                            "local",
		RoutinePoolSize:                     10,
		DBType:                              "postgresql",
		DBUsername:                          "user",
		DBPassword:                          "pass",
		DBHost:                              "localhost",
		DBPort:                              5432,
		DBDatabase:                          "test",
		DBDefaultDatabase:                   "test",
		DBSslMode:                           "disable",
		LifetimeCollectionHeartbeatInterval: 30,
		LifetimeCollectionGCInterval:        300,
		LifetimeStateGCInterval:             60,
		DifyInvocationConnectionIdleTimeout: 300,
		MaxPluginPackageSize:                100 * 1024 * 1024,
		MaxBundlePackageSize:                100 * 1024 * 1024,
		PythonInterpreterPath:               "/usr/bin/python3",
		PythonEnvInitTimeout:                300,
		ClusterDisabled:                     true,
	}

	err := config.Validate()
	assert.NoError(t, err)
}

func TestConfig_GetLocalRuntimeBufferSize_Default(t *testing.T) {
	config := &Config{
		PluginRuntimeBufferSize: 1024,
		PluginStdioBufferSize:   1024,
	}

	assert.Equal(t, 1024, config.GetLocalRuntimeBufferSize())
}

func TestConfig_GetLocalRuntimeBufferSize_CustomStdio(t *testing.T) {
	config := &Config{
		PluginRuntimeBufferSize: 2048,
		PluginStdioBufferSize:   4096, // Custom stdio buffer size
	}

	// Should prefer stdio buffer size when customized
	assert.Equal(t, 4096, config.GetLocalRuntimeBufferSize())
}

func TestConfig_GetLocalRuntimeMaxBufferSize_Default(t *testing.T) {
	config := &Config{
		PluginRuntimeMaxBufferSize: 5242880,
		PluginStdioMaxBufferSize:   5242880,
	}

	assert.Equal(t, 5242880, config.GetLocalRuntimeMaxBufferSize())
}

func TestConfig_GetLocalRuntimeMaxBufferSize_CustomStdio(t *testing.T) {
	config := &Config{
		PluginRuntimeMaxBufferSize: 10485760,
		PluginStdioMaxBufferSize:   20971520, // Custom stdio max buffer size
	}

	// Should prefer stdio max buffer size when customized
	assert.Equal(t, 20971520, config.GetLocalRuntimeMaxBufferSize())
}
