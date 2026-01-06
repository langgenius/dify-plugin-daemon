package local_runtime

import (
	"os"
	"path"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/stretchr/testify/require"
)

func TestDetectDependencyFileType(t *testing.T) {
	t.Run("pyproject.toml exists only", func(t *testing.T) {
		tempDir := t.TempDir()
		pyprojectPath := path.Join(tempDir, "pyproject.toml")
		require.NoError(t, os.WriteFile(pyprojectPath, []byte("[project]\n"), 0644))

		runtime := &LocalPluginRuntime{
			PluginRuntime: plugin_entities.PluginRuntime{
				State: plugin_entities.PluginRuntimeState{
					WorkingPath: tempDir,
				},
			},
		}

		fileType, err := runtime.detectDependencyFileType()
		require.NoError(t, err)
		require.Equal(t, "pyproject.toml", fileType)
	})

	t.Run("requirements.txt exists only", func(t *testing.T) {
		tempDir := t.TempDir()
		requirementsPath := path.Join(tempDir, "requirements.txt")
		require.NoError(t, os.WriteFile(requirementsPath, []byte("dify-plugin==0.1.0\n"), 0644))

		runtime := &LocalPluginRuntime{
			PluginRuntime: plugin_entities.PluginRuntime{
				State: plugin_entities.PluginRuntimeState{
					WorkingPath: tempDir,
				},
			},
		}

		fileType, err := runtime.detectDependencyFileType()
		require.NoError(t, err)
		require.Equal(t, "requirements.txt", fileType)
	})

	t.Run("both files exist - pyproject.toml takes priority", func(t *testing.T) {
		tempDir := t.TempDir()
		pyprojectPath := path.Join(tempDir, "pyproject.toml")
		requirementsPath := path.Join(tempDir, "requirements.txt")
		require.NoError(t, os.WriteFile(pyprojectPath, []byte("[project]\n"), 0644))
		require.NoError(t, os.WriteFile(requirementsPath, []byte("dify-plugin==0.1.0\n"), 0644))

		runtime := &LocalPluginRuntime{
			PluginRuntime: plugin_entities.PluginRuntime{
				State: plugin_entities.PluginRuntimeState{
					WorkingPath: tempDir,
				},
			},
		}

		fileType, err := runtime.detectDependencyFileType()
		require.NoError(t, err)
		require.Equal(t, "pyproject.toml", fileType)
	})

	t.Run("neither file exists", func(t *testing.T) {
		tempDir := t.TempDir()

		runtime := &LocalPluginRuntime{
			PluginRuntime: plugin_entities.PluginRuntime{
				State: plugin_entities.PluginRuntimeState{
					WorkingPath: tempDir,
				},
			},
		}

		fileType, err := runtime.detectDependencyFileType()
		require.Error(t, err)
		require.Empty(t, fileType)
		require.Contains(t, err.Error(), "neither pyproject.toml nor requirements.txt found")
	})
}

func TestGetDependencyFilePath(t *testing.T) {
	t.Run("returns pyproject.toml path when it exists", func(t *testing.T) {
		tempDir := t.TempDir()
		pyprojectPath := path.Join(tempDir, "pyproject.toml")
		require.NoError(t, os.WriteFile(pyprojectPath, []byte("[project]\n"), 0644))

		runtime := &LocalPluginRuntime{
			PluginRuntime: plugin_entities.PluginRuntime{
				State: plugin_entities.PluginRuntimeState{
					WorkingPath: tempDir,
				},
			},
		}

		filePath, err := runtime.getDependencyFilePath()
		require.NoError(t, err)
		require.Equal(t, pyprojectPath, filePath)
	})

	t.Run("returns requirements.txt path when it exists", func(t *testing.T) {
		tempDir := t.TempDir()
		requirementsPath := path.Join(tempDir, "requirements.txt")
		require.NoError(t, os.WriteFile(requirementsPath, []byte("dify-plugin==0.1.0\n"), 0644))

		runtime := &LocalPluginRuntime{
			PluginRuntime: plugin_entities.PluginRuntime{
				State: plugin_entities.PluginRuntimeState{
					WorkingPath: tempDir,
				},
			},
		}

		filePath, err := runtime.getDependencyFilePath()
		require.NoError(t, err)
		require.Equal(t, requirementsPath, filePath)
	})

	t.Run("returns error when no dependency file exists", func(t *testing.T) {
		tempDir := t.TempDir()

		runtime := &LocalPluginRuntime{
			PluginRuntime: plugin_entities.PluginRuntime{
				State: plugin_entities.PluginRuntimeState{
					WorkingPath: tempDir,
				},
			},
		}

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
