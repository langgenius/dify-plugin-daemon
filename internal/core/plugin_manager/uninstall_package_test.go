package plugin_manager

import (
	"os"
	"path/filepath"
	"testing"

	cloudoss "github.com/langgenius/dify-cloud-kit/oss"
	"github.com/langgenius/dify-cloud-kit/oss/factory"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"github.com/stretchr/testify/require"
)

// TestRemoveLocalPluginPackageDeletesPackageFile verifies the fix for
// https://github.com/langgenius/dify-plugin-daemon/issues/797 :
// uninstalling a plugin must also delete the cached .difypkg package file
// from storage/plugin_packages, otherwise it lingers forever.
func TestRemoveLocalPluginPackageDeletesPackageFile(t *testing.T) {
	storageDir := t.TempDir()
	storage, err := factory.Load("local", cloudoss.OSSArgs{
		Local: &cloudoss.Local{Path: storageDir},
	})
	require.NoError(t, err)

	manager := InitGlobalManager(storage, &app.Config{
		Platform:               app.PLATFORM_LOCAL,
		PluginMediaCachePath:   "assets",
		PluginMediaCacheSize:   4,
		PluginAssetCacheSize:   4,
		PluginInstalledPath:    "installed",
		PluginPackageCachePath: "packages",
		PluginWorkingPath:      filepath.Join(storageDir, "working"),
	})

	packageBytes, err := os.ReadFile("testdata/openai.difypkg")
	require.NoError(t, err)
	packageDecoder, err := decoder.NewZipPluginDecoder(packageBytes)
	require.NoError(t, err)
	identifier, err := packageDecoder.UniqueIdentity()
	require.NoError(t, err)

	// seed the cached package file, like an install would
	require.NoError(t, manager.packageBucket.Save(identifier.String(), packageBytes))
	exists, err := manager.packageBucket.Exists(identifier.String())
	require.NoError(t, err)
	require.True(t, exists, "package file should exist before uninstall")

	// act: remove the cached package file (what UninstallPlugin now calls)
	require.NoError(t, manager.RemoveLocalPluginPackage(identifier))

	// assert: the .difypkg file is gone
	exists, err = manager.packageBucket.Exists(identifier.String())
	require.NoError(t, err)
	require.False(t, exists, "package file should be removed after RemoveLocalPluginPackage")
}
