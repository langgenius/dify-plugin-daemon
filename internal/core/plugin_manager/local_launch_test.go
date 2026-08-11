package plugin_manager

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	cloudoss "github.com/langgenius/dify-cloud-kit/oss"
	"github.com/langgenius/dify-cloud-kit/oss/factory"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/installation_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
	"github.com/stretchr/testify/require"
)

func TestEnsureLocalRuntimeFailureKeepsInstalledPackage(t *testing.T) {
	eventData, installed := runFailingLocalRuntimeInstall(t, false)

	require.NotEmpty(t, eventData)
	require.True(t, installed)
}

func TestEnsureLocalRuntimeContinuesWhenRedisLockIsUnavailable(t *testing.T) {
	eventData, installed := runFailingLocalRuntimeInstall(t, true)

	require.Contains(t, eventData, "missing-uv")
	require.NotContains(t, eventData, "failed to acquire distributed env-init lock")
	require.True(t, installed)
}

func runFailingLocalRuntimeInstall(t *testing.T, closeRedisBeforeInstall bool) (string, bool) {
	t.Helper()
	routine.InitPool(4)

	redisServer := miniredis.RunT(t)
	require.NoError(t, cache.InitRedisClient(redisServer.Addr(), cache.RedisCredentials{}, false, 0, nil))
	t.Cleanup(func() {
		_ = cache.Close()
	})

	storageDir := t.TempDir()
	storage, err := factory.Load("local", cloudoss.OSSArgs{
		Local: &cloudoss.Local{Path: storageDir},
	})
	require.NoError(t, err)

	manager := InitGlobalManager(storage, &app.Config{
		Platform:                       app.PLATFORM_LOCAL,
		PluginMediaCachePath:           "assets",
		PluginMediaCacheSize:           4,
		PluginAssetCacheSize:           4,
		PluginInstalledPath:            "installed",
		PluginPackageCachePath:         "packages",
		PluginLocalLaunchingConcurrent: 1,
		PluginWorkingPath:              filepath.Join(storageDir, "working"),
		PythonEnvInitTimeout:           1,
		UvPath:                         filepath.Join(storageDir, "missing-uv"),
	})

	packageBytes, err := os.ReadFile("testdata/openai.difypkg")
	require.NoError(t, err)
	packageDecoder, err := decoder.NewZipPluginDecoder(packageBytes)
	require.NoError(t, err)
	identifier, err := packageDecoder.UniqueIdentity()
	require.NoError(t, err)
	require.NoError(t, manager.packageBucket.Save(identifier.String(), packageBytes))
	if closeRedisBeforeInstall {
		require.NoError(t, cache.Close())
	}

	response, err := manager.EnsureRuntime(context.Background(), identifier)
	require.NoError(t, err)
	eventData := ""
	require.NoError(t, response.Process(func(event installation_entities.PluginInstallResponse) {
		if event.Event == installation_entities.PluginInstallEventError {
			eventData = event.Data
		}
	}))

	exists, err := manager.installedBucket.Exists(identifier)
	require.NoError(t, err)
	return eventData, exists
}
