package local_runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/constants"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/stretchr/testify/require"
)

func pluginEnvSliceToMap(t *testing.T, env []string) map[string]string {
	t.Helper()
	envByKey := make(map[string]string, len(env))
	for _, item := range env {
		key, value, found := strings.Cut(item, "=")
		require.True(t, found)
		envByKey[key] = value
	}
	return envByKey
}

func TestBuildPluginCommandEnv(t *testing.T) {
	t.Setenv("PATH", "/test/bin")
	t.Setenv("HOME", "/test/home")
	t.Setenv("LANG", "en_US.UTF-8")
	t.Setenv("HTTP_PROXY", "http://env-proxy:8080")
	t.Setenv("DB_PASSWORD", "must-not-be-inherited")
	t.Setenv("SERVER_KEY", "must-not-be-inherited")
	t.Setenv("DIFY_INNER_API_KEY", "must-not-be-inherited")
	t.Setenv("AWS_ACCESS_KEY_ID", "must-not-be-inherited")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "must-not-be-inherited")
	t.Setenv("REDIS_PASSWORD", "must-not-be-inherited")
	t.Setenv("ADMIN_API_KEY", "must-not-be-inherited")
	t.Setenv("UNRELATED_SECRET", "must-not-be-inherited")

	env := BuildPluginCommandEnv(&app.Config{
		HttpsProxy: "https://config-proxy:8443",
		NoProxy:    "localhost,127.0.0.1",
	})
	envByKey := pluginEnvSliceToMap(t, env)

	// allowlisted variables are passed through
	require.Equal(t, "/test/bin", envByKey["PATH"])
	require.Equal(t, "/test/home", envByKey["HOME"])
	require.Equal(t, "en_US.UTF-8", envByKey["LANG"])
	require.Equal(t, "http://env-proxy:8080", envByKey["HTTP_PROXY"])
	require.Equal(t, "local", envByKey["INSTALL_METHOD"])

	// proxy settings from the daemon config take precedence
	require.Equal(t, "https://config-proxy:8443", envByKey["HTTPS_PROXY"])
	require.Equal(t, "localhost,127.0.0.1", envByKey["NO_PROXY"])

	// daemon secrets never reach the plugin process
	for _, key := range []string{
		"DB_PASSWORD",
		"SERVER_KEY",
		"DIFY_INNER_API_KEY",
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"REDIS_PASSWORD",
		"ADMIN_API_KEY",
		"UNRELATED_SECRET",
	} {
		require.NotContains(t, envByKey, key)
	}
}

func TestGetInstanceCmdDoesNotInheritDaemonEnv(t *testing.T) {
	t.Setenv("PATH", "/test/bin")
	t.Setenv("DB_PASSWORD", "must-not-be-inherited")
	t.Setenv("SERVER_KEY", "must-not-be-inherited")
	t.Setenv("DIFY_INNER_API_KEY", "must-not-be-inherited")

	// minimal fake venv so getVirtualEnvironmentPythonPath succeeds
	workDir := t.TempDir()
	venvBin := filepath.Join(workDir, ".venv", "bin")
	require.NoError(t, os.MkdirAll(venvBin, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(venvBin, "python"), []byte("#!/bin/sh\n"), 0o755))

	r := &LocalPluginRuntime{appConfig: &app.Config{}}
	r.State = plugin_entities.PluginRuntimeState{WorkingPath: workDir}
	r.Config.Meta.Runner.Language = constants.Python
	r.Config.Meta.Runner.Entrypoint = "main"

	cmd, err := r.getInstanceCmd()
	require.NoError(t, err)

	envByKey := pluginEnvSliceToMap(t, cmd.Env)
	require.Equal(t, "/test/bin", envByKey["PATH"])
	require.Equal(t, "local", envByKey["INSTALL_METHOD"])
	require.NotContains(t, envByKey, "DB_PASSWORD")
	require.NotContains(t, envByKey, "SERVER_KEY")
	require.NotContains(t, envByKey, "DIFY_INNER_API_KEY")
}
