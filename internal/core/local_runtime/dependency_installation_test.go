package local_runtime

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"github.com/stretchr/testify/require"
)

func copyTestData(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(targetPath, data, info.Mode())
	})
	require.NoError(t, err)
}

func createTestRuntime(t *testing.T, pluginDir string) *LocalPluginRuntime {
	t.Helper()
	tempDir := t.TempDir()
	pluginSourceDir := path.Join("testdata", pluginDir)

	pluginDecoder, err := decoder.NewFSPluginDecoder(pluginSourceDir)
	require.NoError(t, err)

	appConfig := &app.Config{
		PluginWorkingPath: tempDir,
	}

	runtime, err := ConstructPluginRuntime(appConfig, pluginDecoder)
	require.NoError(t, err)

	copyTestData(t, pluginSourceDir, runtime.State.WorkingPath)

	return runtime
}

func TestDetectDependencyFileType(t *testing.T) {
	t.Run("pyproject.toml exists only", func(t *testing.T) {
		runtime := createTestRuntime(t, "plugin-with-pyproject")

		fileType, err := runtime.detectDependencyFileType()
		require.NoError(t, err)
		require.Equal(t, pyprojectTomlFile, fileType)
	})

	t.Run("requirements.txt exists only", func(t *testing.T) {
		runtime := createTestRuntime(t, "plugin-with-requirements")

		fileType, err := runtime.detectDependencyFileType()
		require.NoError(t, err)
		require.Equal(t, requirementsTxtFile, fileType)
	})

	t.Run("both files exist - pyproject.toml takes priority", func(t *testing.T) {
		runtime := createTestRuntime(t, "plugin-with-both")

		fileType, err := runtime.detectDependencyFileType()
		require.NoError(t, err)
		require.Equal(t, pyprojectTomlFile, fileType)
	})

	t.Run("neither file exists", func(t *testing.T) {
		runtime := createTestRuntime(t, "plugin-without-dependencies")

		fileType, err := runtime.detectDependencyFileType()
		require.Error(t, err)
		require.Empty(t, fileType)
		require.Contains(t, err.Error(), "neither pyproject.toml nor requirements.txt found")
	})
}

func TestGetDependencyFilePath(t *testing.T) {
	t.Run("returns pyproject.toml path when it exists", func(t *testing.T) {
		runtime := createTestRuntime(t, "plugin-with-pyproject")

		filePath, err := runtime.getDependencyFilePath()
		require.NoError(t, err)
		require.Equal(t, path.Join(runtime.State.WorkingPath, "pyproject.toml"), filePath)
	})

	t.Run("returns requirements.txt path when it exists", func(t *testing.T) {
		runtime := createTestRuntime(t, "plugin-with-requirements")

		filePath, err := runtime.getDependencyFilePath()
		require.NoError(t, err)
		require.Equal(t, path.Join(runtime.State.WorkingPath, "requirements.txt"), filePath)
	})

	t.Run("returns error when no dependency file exists", func(t *testing.T) {
		runtime := createTestRuntime(t, "plugin-without-dependencies")

		filePath, err := runtime.getDependencyFilePath()
		require.Error(t, err)
		require.Empty(t, filePath)
	})
}

func TestPrepareSyncArgs(t *testing.T) {
	t.Run("basic sync args without config", func(t *testing.T) {
		runtime := &LocalPluginRuntime{
			appConfig: &app.Config{},
		}

		args := runtime.prepareSyncArgs()
		require.Equal(t, []string{"sync", "--no-dev"}, args)
	})

	t.Run("sync args with mirror URL", func(t *testing.T) {
		runtime := &LocalPluginRuntime{
			appConfig: &app.Config{
				PipMirrorUrl: "https://pypi.tuna.tsinghua.edu.cn/simple",
			},
		}

		args := runtime.prepareSyncArgs()
		require.Equal(t, []string{"sync", "--no-dev", "-i", "https://pypi.tuna.tsinghua.edu.cn/simple"}, args)
	})

	t.Run("sync args with verbose flag", func(t *testing.T) {
		runtime := &LocalPluginRuntime{
			appConfig: &app.Config{
				PipVerbose: true,
			},
		}

		args := runtime.prepareSyncArgs()
		require.Equal(t, []string{"sync", "--no-dev", "-v"}, args)
	})

	t.Run("sync args with extra args", func(t *testing.T) {
		runtime := &LocalPluginRuntime{
			appConfig: &app.Config{
				PipExtraArgs: "--no-cache --retries 3",
			},
		}

		args := runtime.prepareSyncArgs()
		require.Equal(t, []string{"sync", "--no-dev", "--no-cache", "--retries", "3"}, args)
	})

	t.Run("sync args with all config options", func(t *testing.T) {
		runtime := &LocalPluginRuntime{
			appConfig: &app.Config{
				PipMirrorUrl: "https://pypi.tuna.tsinghua.edu.cn/simple",
				PipVerbose:   true,
				PipExtraArgs: "--no-cache",
			},
		}

		args := runtime.prepareSyncArgs()
		require.Equal(t, []string{
			"sync",
			"--no-dev",
			"-i", "https://pypi.tuna.tsinghua.edu.cn/simple",
			"-v",
			"--no-cache",
		}, args)
	})
}

func TestPreparePipArgs(t *testing.T) {
	t.Run("basic pip args without config", func(t *testing.T) {
		runtime := &LocalPluginRuntime{
			appConfig: &app.Config{},
		}

		args := runtime.preparePipArgs()
		require.Equal(t, []string{"pip", "install", "-r", "requirements.txt"}, args)
	})

	t.Run("pip args with mirror URL", func(t *testing.T) {
		runtime := &LocalPluginRuntime{
			appConfig: &app.Config{
				PipMirrorUrl: "https://pypi.tuna.tsinghua.edu.cn/simple",
			},
		}

		args := runtime.preparePipArgs()
		require.Equal(t, []string{
			"pip",
			"install",
			"-i", "https://pypi.tuna.tsinghua.edu.cn/simple",
			"--trusted-host", "pypi.tuna.tsinghua.edu.cn",
			"-r", "requirements.txt",
		}, args)
	})

	t.Run("pip args with verbose flag", func(t *testing.T) {
		runtime := &LocalPluginRuntime{
			appConfig: &app.Config{
				PipVerbose: true,
			},
		}

		args := runtime.preparePipArgs()
		require.Equal(t, []string{"pip", "install", "-r", "requirements.txt", "-vvv"}, args)
	})

	t.Run("pip args with all config options", func(t *testing.T) {
		runtime := &LocalPluginRuntime{
			appConfig: &app.Config{
				PipMirrorUrl: "https://pypi.tuna.tsinghua.edu.cn/simple",
				PipVerbose:   true,
				PipExtraArgs: "--no-cache",
			},
		}

		args := runtime.preparePipArgs()
		require.Equal(t, []string{
			"pip",
			"install",
			"-i", "https://pypi.tuna.tsinghua.edu.cn/simple",
			"--trusted-host", "pypi.tuna.tsinghua.edu.cn",
			"-r", "requirements.txt",
			"-vvv",
			"--no-cache",
		}, args)
	})
}

func TestGetPluginSdkVersionWithPyprojectToml(t *testing.T) {
	runtime := &LocalPluginRuntime{}

	t.Run("pyproject.toml with exact version in dependencies", func(t *testing.T) {
		pyprojectContent := `
[project]
name = "test-plugin"
dependencies = [
    "dify-plugin==0.1.5",
    "pydantic>=2.0.0",
]
`
		version, err := runtime.getPluginSdkVersion(pyprojectContent)
		require.NoError(t, err)
		require.Equal(t, "0.1.5", version)
	})

	t.Run("pyproject.toml with version range", func(t *testing.T) {
		pyprojectContent := `
[project]
dependencies = [
    "dify-plugin>=0.1.0,<0.2.0",
]
`
		version, err := runtime.getPluginSdkVersion(pyprojectContent)
		require.NoError(t, err)
		require.Equal(t, "0.2.0", version)
	})

	t.Run("pyproject.toml with underscore in package name", func(t *testing.T) {
		pyprojectContent := `
[project]
dependencies = ["dify_plugin==0.1.3"]
`
		version, err := runtime.getPluginSdkVersion(pyprojectContent)
		require.NoError(t, err)
		require.Equal(t, "0.1.3", version)
	})

	t.Run("pyproject.toml with compatible version", func(t *testing.T) {
		pyprojectContent := `
[project]
dependencies = [
    "dify-plugin~=0.1.0",
]
`
		version, err := runtime.getPluginSdkVersion(pyprojectContent)
		require.NoError(t, err)
		require.Equal(t, "0.1.0", version)
	})

	t.Run("pyproject.toml without dify-plugin", func(t *testing.T) {
		pyprojectContent := `
[project]
dependencies = [
    "pydantic>=2.0.0",
]
`
		_, err := runtime.getPluginSdkVersion(pyprojectContent)
		require.Error(t, err)
	})
}
