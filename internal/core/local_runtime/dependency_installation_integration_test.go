package local_runtime

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
	"github.com/stretchr/testify/require"
)

func TestInitPythonEnvironmentWithPyprojectToml(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	routine.InitPool(1024)

	uvPath := findUVPath(t)
	pythonPath := findPythonPath(t)

	testCases := []struct {
		name                  string
		pluginDir             string
		expectedDependency    PythonDependencyFileType
		shouldPreferPyproject bool
		shouldFail            bool
	}{
		{
			name:                  "plugin with pyproject.toml only",
			pluginDir:             "plugin-with-pyproject",
			expectedDependency:    pyprojectTomlFile,
			shouldPreferPyproject: true,
			shouldFail:            false,
		},
		{
			name:                  "plugin with requirements.txt only",
			pluginDir:             "plugin-with-requirements",
			expectedDependency:    requirementsTxtFile,
			shouldPreferPyproject: false,
			shouldFail:            false,
		},
		{
			name:                  "plugin with both files - pyproject.toml takes priority",
			pluginDir:             "plugin-with-both",
			expectedDependency:    pyprojectTomlFile,
			shouldPreferPyproject: true,
			shouldFail:            false,
		},
		{
			name:                  "plugin without any dependency file - should fail",
			pluginDir:             "plugin-without-dependencies",
			expectedDependency:    "",
			shouldPreferPyproject: false,
			shouldFail:            true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()
			pluginSourceDir := path.Join("testdata", tc.pluginDir)

			// Create decoder from source directory
			pluginDecoder, err := decoder.NewFSPluginDecoder(pluginSourceDir)
			require.NoError(t, err)

			appConfig := &app.Config{
				PythonInterpreterPath: pythonPath,
				UvPath:                uvPath,
				PythonEnvInitTimeout:  120,
				PluginWorkingPath:     tempDir,
			}

			// Use ConstructPluginRuntime to properly initialize all fields
			runtime, err := ConstructPluginRuntime(appConfig, pluginDecoder)
			require.NoError(t, err)

			// Copy plugin files to the computed working path
			require.NoError(t, copyDir(pluginSourceDir, runtime.State.WorkingPath))

			t.Logf("Testing plugin in: %s", runtime.State.WorkingPath)

			if tc.shouldFail {
				// Test that file detection fails
				fileType, err := runtime.detectDependencyFileType()
				require.Error(t, err, "detectDependencyFileType should fail when no dependency file exists")
				require.Empty(t, fileType)
				require.Contains(t, err.Error(), "neither pyproject.toml nor requirements.txt found")

				// Test that InitPythonEnvironment fails gracefully
				err = runtime.InitPythonEnvironment()
				require.Error(t, err, "InitPythonEnvironment should fail when no dependency file exists")
				require.Contains(t, err.Error(), "failed to create virtual environment")

				t.Logf("Correctly failed with error: %v", err)
			} else {
				// Test successful case
				fileType, err := runtime.detectDependencyFileType()
				require.NoError(t, err)
				require.Equal(t, tc.expectedDependency, fileType)

				err = runtime.InitPythonEnvironment()
				require.NoError(t, err, "InitPythonEnvironment should succeed")

				venvPath := path.Join(runtime.State.WorkingPath, ".venv")
				require.DirExists(t, venvPath, "Virtual environment should be created")

				pythonBinPath := path.Join(venvPath, "bin", "python")
				require.FileExists(t, pythonBinPath, "Python binary should exist in venv")

				validFlagPath := path.Join(venvPath, "dify", "plugin.json")
				require.FileExists(t, validFlagPath, "Valid flag file should exist")

				cmd := exec.Command(pythonBinPath, "-c", "import dify_plugin; print('SUCCESS')")
				cmd.Dir = runtime.State.WorkingPath
				output, err := cmd.CombinedOutput()
				if err != nil {
					t.Logf("Python import failed. Output: %s", string(output))
				}
				require.NoError(t, err, "Should be able to import dify_plugin")
				require.Contains(t, string(output), "SUCCESS", "dify_plugin should import successfully")
				t.Logf("dify_plugin imported successfully")

				if tc.shouldPreferPyproject {
					lockfilePath := path.Join(runtime.State.WorkingPath, "uv.lock")
					t.Logf("Checking for uv.lock at: %s", lockfilePath)
					if _, err := os.Stat(lockfilePath); err == nil {
						t.Logf("uv.lock exists (uv sync was used)")
					} else {
						t.Logf("uv.lock does not exist (may not be created in all cases)")
					}
				}
			}
		})
	}
}

func TestInitPythonEnvironmentErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	uvPath := findUVPath(t)
	pythonPath := findPythonPath(t)

	t.Run("fails when no dependency file exists", func(t *testing.T) {
		tempDir := t.TempDir()
		pluginSourceDir := path.Join("testdata", "plugin-without-dependencies")

		pluginDecoder, err := decoder.NewFSPluginDecoder(pluginSourceDir)
		require.NoError(t, err)

		appConfig := &app.Config{
			PythonInterpreterPath: pythonPath,
			UvPath:                uvPath,
			PythonEnvInitTimeout:  120,
			PluginWorkingPath:     tempDir,
		}

		runtime, err := ConstructPluginRuntime(appConfig, pluginDecoder)
		require.NoError(t, err)

		require.NoError(t, copyDir(pluginSourceDir, runtime.State.WorkingPath))

		_, err = runtime.detectDependencyFileType()
		require.Error(t, err)
		require.Contains(t, err.Error(), "neither pyproject.toml nor requirements.txt found")
	})
}

func TestCreateVirtualEnvironmentValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	uvPath := findUVPath(t)
	pythonPath := findPythonPath(t)

	t.Run("validates pyproject.toml exists", func(t *testing.T) {
		tempDir := t.TempDir()
		pluginSourceDir := path.Join("testdata", "plugin-with-pyproject")

		pluginDecoder, err := decoder.NewFSPluginDecoder(pluginSourceDir)
		require.NoError(t, err)

		appConfig := &app.Config{
			PythonInterpreterPath: pythonPath,
			UvPath:                uvPath,
			PythonEnvInitTimeout:  120,
			PluginWorkingPath:     tempDir,
		}

		runtime, err := ConstructPluginRuntime(appConfig, pluginDecoder)
		require.NoError(t, err)

		require.NoError(t, copyDir(pluginSourceDir, runtime.State.WorkingPath))

		venv, err := runtime.createVirtualEnvironment(uvPath)
		require.NoError(t, err, "Should create venv when pyproject.toml exists")
		require.NotNil(t, venv)
	})

	t.Run("validates requirements.txt exists", func(t *testing.T) {
		tempDir := t.TempDir()
		pluginSourceDir := path.Join("testdata", "plugin-with-requirements")

		pluginDecoder, err := decoder.NewFSPluginDecoder(pluginSourceDir)
		require.NoError(t, err)

		appConfig := &app.Config{
			PythonInterpreterPath: pythonPath,
			UvPath:                uvPath,
			PythonEnvInitTimeout:  120,
			PluginWorkingPath:     tempDir,
		}

		runtime, err := ConstructPluginRuntime(appConfig, pluginDecoder)
		require.NoError(t, err)

		require.NoError(t, copyDir(pluginSourceDir, runtime.State.WorkingPath))

		venv, err := runtime.createVirtualEnvironment(uvPath)
		require.NoError(t, err, "Should create venv when requirements.txt exists")
		require.NotNil(t, venv)
	})
}

func findUVPath(t *testing.T) string {
	t.Helper()

	uvPath := os.Getenv("UV_PATH")
	if uvPath != "" {
		return uvPath
	}

	cmd := exec.Command("which", "uv")
	output, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("python3", "-c", "from uv._find_uv import find_uv_bin; print(find_uv_bin())")
		output, err = cmd.Output()
		if err != nil {
			t.Fatal("UV not found. Install UV to run integration tests: https://docs.astral.sh/uv/")
			return ""
		}
	}

	uvPath = strings.TrimSpace(string(output))
	if uvPath == "" {
		t.Fatal("UV not found. Install UV to run integration tests: https://docs.astral.sh/uv/")
	}

	t.Logf("Found UV at: %s", uvPath)
	return uvPath
}

func findPythonPath(t *testing.T) string {
	t.Helper()

	pythonPath := os.Getenv("PYTHON_INTERPRETER_PATH")
	if pythonPath != "" {
		return pythonPath
	}

	for _, name := range []string{"python3.12", "python3.11", "python3.10", "python3"} {
		cmd := exec.Command("which", name)
		output, err := cmd.Output()
		if err == nil {
			pythonPath = strings.TrimSpace(string(output))
			if pythonPath != "" {
				t.Logf("Found Python at: %s", pythonPath)
				return pythonPath
			}
		}
	}

	t.Fatal("Python 3 not found. Install Python 3.10+ to run integration tests")
	return ""
}

func copyDir(src string, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		return os.WriteFile(targetPath, data, info.Mode())
	})
}
