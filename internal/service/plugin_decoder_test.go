package service

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	cloudoss "github.com/langgenius/dify-cloud-kit/oss"
	"github.com/langgenius/dify-cloud-kit/oss/factory"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/signer/withkey"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache/helper"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/encryption"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type uploadTestEnv struct {
	config     *app.Config
	manager    *plugin_manager.PluginManager
	storageDir string
}

type testMultipartFile struct {
	*bytes.Reader
}

func (f testMultipartFile) Close() error {
	return nil
}

func setupUploadTestEnv(t *testing.T, forceVerify bool) uploadTestEnv {
	t.Helper()

	redisServer := miniredis.RunT(t)
	require.NoError(t, cache.InitRedisClient(redisServer.Addr(), "", "", false, 0, nil))
	t.Cleanup(func() {
		_ = cache.Close()
	})

	gormDB, err := gorm.Open(sqlite.Open("file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, gormDB.AutoMigrate(
		&models.Plugin{},
		&models.PluginDeclaration{},
		&models.PluginInstallation{},
	))
	db.DifyPluginDB = gormDB
	t.Cleanup(func() {
		sqlDB, err := gormDB.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	storageDir := t.TempDir()
	oss, err := factory.Load("local", cloudoss.OSSArgs{
		Local: &cloudoss.Local{Path: storageDir},
	})
	require.NoError(t, err)

	config := &app.Config{
		PluginMediaCachePath:                   "assets",
		PluginMediaCacheSize:                   32,
		PluginAssetCacheSize:                   32,
		PluginInstalledPath:                    "installed",
		PluginPackageCachePath:                 "packages",
		PluginLocalLaunchingConcurrent:         1,
		PluginInstallTimeout:                   15,
		PluginWorkingPath:                      filepath.Join(storageDir, "working"),
		Platform:                               app.PLATFORM_LOCAL,
		MaxPluginPackageSize:                   50 * 1024 * 1024,
		MaxBundlePackageSize:                   50 * 1024 * 1024,
		ForceVerifyingSignature:                forceVerify,
		EnforceLanggeniusSignatures:            false,
		ThirdPartySignatureVerificationEnabled: true,
		ThirdPartySignatureVerificationPublicKeys: []string{
			repoPath(t, "pkg/plugin_packager/testdata/keys/test_key_pair_1.public.pem"),
		},
	}

	manager := plugin_manager.InitGlobalManager(oss, config)
	return uploadTestEnv{config: config, manager: manager, storageDir: storageDir}
}

func TestUploadPluginPkgRejectsInvalidSignatureWithoutArtifacts(t *testing.T) {
	env := setupUploadTestEnv(t, true)
	pkgBytes := signedPluginPackage(t, "badpkgsingle")
	pkgBytes[len(pkgBytes)-1] = 0
	pkgBytes[len(pkgBytes)-2] = 0
	identifier := packageIdentifier(t, pkgBytes)

	resp := UploadPluginPkg(env.config, nil, "tenant-1", testMultipartFile{bytes.NewReader(pkgBytes)}, true)

	require.Equal(t, -400, resp.Code)
	requireNoUploadArtifacts(t, env, identifier)
	require.Empty(t, storedFiles(t, filepath.Join(env.storageDir, "assets")))
}

func TestUploadPluginBundleRejectsInvalidPackageDependencyWithoutPartialArtifacts(t *testing.T) {
	env := setupUploadTestEnv(t, true)
	validDependency := signedPluginPackage(t, "bundlevaliddep")
	invalidDependency := signedPluginPackage(t, "bundlebaddep")
	invalidDependency[len(invalidDependency)-1] = 0
	invalidDependency[len(invalidDependency)-2] = 0

	validIdentifier := packageIdentifier(t, validDependency)
	invalidIdentifier := packageIdentifier(t, invalidDependency)
	bundleBytes := bundleWithPackageDependencies(t, []bundlePackage{
		{name: "valid.difypkg", content: validDependency},
		{name: "bad.difypkg", content: invalidDependency},
	})

	resp := UploadPluginBundle(env.config, nil, "tenant-1", testMultipartFile{bytes.NewReader(bundleBytes)}, true)

	require.NotEqual(t, 0, resp.Code)
	requireNoUploadArtifacts(t, env, validIdentifier, invalidIdentifier)
	require.Empty(t, storedFiles(t, filepath.Join(env.storageDir, "assets")))
}

func TestUploadPluginPkgPersistsValidPackageAndSupportsInstallLookup(t *testing.T) {
	env := setupUploadTestEnv(t, true)
	pkgBytes := signedPluginPackage(t, "validpkgupload")
	identifier := packageIdentifier(t, pkgBytes)

	resp := UploadPluginPkg(env.config, nil, "tenant-1", testMultipartFile{bytes.NewReader(pkgBytes)}, true)

	require.Equal(t, 0, resp.Code)
	storedPackage, err := env.manager.GetPackage(identifier)
	require.NoError(t, err)
	require.Equal(t, pkgBytes, storedPackage)

	manifestResp := FetchPluginManifest(identifier)
	require.Equal(t, 0, manifestResp.Code)

	declaration := manifestResp.Data.(*plugin_entities.PluginDeclaration)
	require.True(t, declaration.Verified)
	require.Equal(t, "validpkgupload", declaration.Name)

	require.NoError(t, db.Create(&models.Plugin{
		PluginUniqueIdentifier: identifier.String(),
		PluginID:               identifier.PluginID(),
		InstallType:            plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL,
	}))
	installResp := InstallMultiplePluginsToTenant(
		context.Background(),
		env.config,
		"00000000-0000-0000-0000-000000000001",
		[]plugin_entities.PluginUniqueIdentifier{identifier},
		"test",
		[]map[string]any{{}},
	)
	require.Equal(t, 0, installResp.Code)
	installData := installResp.Data.(*InstallPluginResponse)
	require.True(t, installData.AllInstalled)
	require.Empty(t, installData.TaskID)
}

func TestUploadPluginPkgStillAllowsUnsignedPackageWhenVerificationNotRequired(t *testing.T) {
	env := setupUploadTestEnv(t, false)
	pkgBytes := unsignedPluginPackage(t, "unsignedallowed")
	identifier := packageIdentifier(t, pkgBytes)

	resp := UploadPluginPkg(env.config, nil, "tenant-1", testMultipartFile{bytes.NewReader(pkgBytes)}, false)

	require.Equal(t, 0, resp.Code)
	_, err := env.manager.GetPackage(identifier)
	require.NoError(t, err)
	manifestResp := FetchPluginManifest(identifier)
	require.Equal(t, 0, manifestResp.Code)
}

func requireNoUploadArtifacts(t *testing.T, env uploadTestEnv, identifiers ...plugin_entities.PluginUniqueIdentifier) {
	t.Helper()

	for _, identifier := range identifiers {
		_, err := env.manager.GetPackage(identifier)
		require.Error(t, err)

		_, err = db.GetOne[models.PluginDeclaration](
			db.Equal("plugin_unique_identifier", identifier.String()),
		)
		require.ErrorIs(t, err, db.ErrDatabaseNotFound)

		resp := FetchPluginManifest(identifier)
		require.Equal(t, -400, resp.Code)
		require.Contains(t, resp.Message, helper.ErrPluginNotFound.Error())
	}
}

func signedPluginPackage(t *testing.T, name string) []byte {
	t.Helper()

	pkgBytes := unsignedPluginPackage(t, name)
	privateKeyBytes, err := os.ReadFile(repoPath(t, "pkg/plugin_packager/testdata/keys/test_key_pair_1.private.pem"))
	require.NoError(t, err)
	privateKey, err := encryption.LoadPrivateKey(privateKeyBytes)
	require.NoError(t, err)
	signed, err := withkey.SignPluginWithPrivateKey(pkgBytes, &decoder.Verification{
		AuthorizedCategory: decoder.AUTHORIZED_CATEGORY_LANGGENIUS,
	}, privateKey)
	require.NoError(t, err)
	return signed
}

func unsignedPluginPackage(t *testing.T, name string) []byte {
	t.Helper()

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)
	writeZipFile(t, zipWriter, "manifest.yaml", pluginManifest(name))
	writeZipFile(t, zipWriter, "endpoint.yaml", endpointManifest())
	writeZipFile(t, zipWriter, "_assets/test.svg", `<svg xmlns="http://www.w3.org/2000/svg"></svg>`)
	require.NoError(t, zipWriter.Close())
	return buf.Bytes()
}

type bundlePackage struct {
	name    string
	content []byte
}

func bundleWithPackageDependencies(t *testing.T, packages []bundlePackage) []byte {
	t.Helper()

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)
	writeZipFile(t, zipWriter, "manifest.yaml", bundleManifest(packages))
	for _, pkg := range packages {
		writeZipFileBytes(t, zipWriter, filepath.Join("_assets", pkg.name), pkg.content)
	}
	require.NoError(t, zipWriter.Close())
	return buf.Bytes()
}

func packageIdentifier(t *testing.T, pkgBytes []byte) plugin_entities.PluginUniqueIdentifier {
	t.Helper()

	zipDecoder, err := decoder.NewZipPluginDecoder(pkgBytes)
	require.NoError(t, err)
	identifier, err := zipDecoder.UniqueIdentity()
	require.NoError(t, err)
	return identifier
}

func writeZipFile(t *testing.T, zipWriter *zip.Writer, name string, content string) {
	t.Helper()
	writeZipFileBytes(t, zipWriter, name, []byte(content))
}

func writeZipFileBytes(t *testing.T, zipWriter *zip.Writer, name string, content []byte) {
	t.Helper()
	file, err := zipWriter.Create(name)
	require.NoError(t, err)
	_, err = io.Copy(file, bytes.NewReader(content))
	require.NoError(t, err)
}

func storedFiles(t *testing.T, root string) []string {
	t.Helper()

	files := []string{}
	if _, err := os.Stat(root); errors.Is(err, os.ErrNotExist) {
		return files
	}
	require.NoError(t, filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, rel)
		return nil
	}))
	return files
}

func pluginManifest(name string) string {
	return fmt.Sprintf(`version: 0.0.1
type: plugin
author: testauthor
name: %s
icon: test.svg
description:
  en_US: "test"
label:
  en_US: "Test"
created_at: "2024-07-12T08:03:44Z"
resource:
  memory: 1048576
  permission:
    endpoint:
      enabled: true
plugins:
  endpoints:
    - "endpoint.yaml"
meta:
  version: 0.0.1
  arch:
    - "amd64"
  runner:
    language: "python"
    version: "3.12"
    entrypoint: "main.py"
`, name)
}

func endpointManifest() string {
	return `settings: []
endpoints:
  - path: "/test"
    method: "GET"
`
}

func bundleManifest(packages []bundlePackage) string {
	var builder strings.Builder
	builder.WriteString(`version: 0.0.1
name: test-bundle
labels:
  en_US: "Test Bundle"
description:
  en_US: "Test bundle"
icon: test.svg
author: testauthor
type: bundle
dependencies:
`)
	for _, pkg := range packages {
		builder.WriteString("  - type: package\n")
		builder.WriteString("    value:\n")
		builder.WriteString("      path: " + pkg.name + "\n")
	}
	return builder.String()
}

func repoPath(t *testing.T, parts ...string) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	return filepath.Join(append([]string{root}, parts...)...)
}
