package local_runtime

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
	"github.com/stretchr/testify/require"
)

func envSliceToMap(in []string) map[string]string {
	m := map[string]string{}
	for _, kv := range in {
		if kv == "" {
			continue
		}
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			continue
		}
		m[parts[0]] = parts[1]
	}
	return m
}

func TestSplitByCommaOrSpace(t *testing.T) {
	cases := []struct {
		in  string
		out []string
	}{
		{"", []string{}},
		{"a", []string{"a"}},
		{"a,b", []string{"a", "b"}},
		{"a b", []string{"a", "b"}},
		{" a, b  c ", []string{"a", "b", "c"}},
		{"https://a/simple, https://b/simple", []string{"https://a/simple", "https://b/simple"}},
	}
	for _, c := range cases {
		got := splitByCommaOrSpace(c.in)
		require.Equal(t, c.out, got)
	}
}

func TestSanitizeArgs(t *testing.T) {
	in := []string{"sync", "-i", "https://user:pass@example.com/simple", "--extra-index-url", "https://token@mirror.example/simple"}
	got := sanitizeArgs(in)
	joined := strings.Join(got, " ")
	require.NotContains(t, joined, "user:pass@")
	require.NotContains(t, joined, "token@")
}

func TestPrepareSyncArgs_IndexAndExtras(t *testing.T) {
	// Uses only PIP_MIRROR_URL and PIP_EXTRA_INDEX_URL
	r := &LocalPluginRuntime{appConfig: &app.Config{PipMirrorUrl: "https://mirror.example/simple"}}
	args := r.prepareSyncArgs()
	require.Equal(t, []string{"sync", "--no-dev", "-i", "https://mirror.example/simple"}, args)
}

func TestPreparePipArgs_IndexAndExtras(t *testing.T) {
	// Index from PIP_MIRROR_URL
	r := &LocalPluginRuntime{appConfig: &app.Config{PipMirrorUrl: "https://mirror.example/simple"}}
	args := r.preparePipArgs()
	joined := strings.Join(args, " ")
	require.Contains(t, joined, "-i https://mirror.example/simple")
	require.Contains(t, joined, "--trusted-host mirror.example")

	// Extra indexes from PIP_EXTRA_INDEX_URL
	r = &LocalPluginRuntime{appConfig: &app.Config{PipExtraIndexUrl: "https://a/simple https://b/simple"}}
	args = r.preparePipArgs()
	joined = strings.Join(args, " ")
	require.Contains(t, joined, "--extra-index-url https://a/simple")
	require.Contains(t, joined, "--extra-index-url https://b/simple")
	require.Contains(t, joined, "--trusted-host a")
	require.Contains(t, joined, "--trusted-host b")
}

// writeFakeUv creates a uv shim that records args and simulates `uv venv` by
// creating the expected virtualenv layout under the current working directory.
func writeFakeUv(t *testing.T, recordFile string) string {
	t.Helper()
	script := "#!/usr/bin/env bash\n" +
		"set -euo pipefail\n" +
		"echo \"$0 $@\" >> \"" + recordFile + "\"\n" +
		"if [[ \"${1:-}\" == venv ]]; then\n" +
		"  envdir=\"${2:-.venv}\"\n" +
		"  mkdir -p \"$envdir/bin\"\n" +
		"  : > \"$envdir/bin/python\"\n" +
		"  chmod +x \"$envdir/bin/python\"\n" +
		"  exit 0\n" +
		"fi\n" +
		"exit 0\n"
	p := path.Join(t.TempDir(), "uv")
	require.NoError(t, os.WriteFile(p, []byte(script), 0755))
	return p
}

func TestInstallDependencies_UvSync_WithIndexArgs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	routine.InitPool(64)
	record := path.Join(t.TempDir(), "uv_args.log")
	fakeUv := writeFakeUv(t, record)

	pluginSourceDir := path.Join("testdata", "plugin-with-pyproject")
	dec, err := decoder.NewFSPluginDecoder(pluginSourceDir)
	require.NoError(t, err)

	appCfg := &app.Config{UvPath: fakeUv, PythonEnvInitTimeout: 60, PipMirrorUrl: "https://uv.example/simple", PipExtraIndexUrl: "https://a/simple https://b/simple"}
	runtime, err := ConstructPluginRuntime(appCfg, dec)
	require.NoError(t, err)
	require.NoError(t, copyDir(pluginSourceDir, runtime.State.WorkingPath))
	_, err = runtime.createVirtualEnvironment(fakeUv)
	require.NoError(t, err)
	depType, err := runtime.detectDependencyFileType()
	require.NoError(t, err)
	require.NoError(t, runtime.installDependencies(fakeUv, depType))

	b, err := os.ReadFile(record)
	require.NoError(t, err)
	out := string(b)
	require.Contains(t, out, "sync --no-dev -i https://uv.example/simple")
	require.Contains(t, out, "--extra-index-url https://a/simple")
	require.Contains(t, out, "--extra-index-url https://b/simple")
}

func TestInstallDependencies_UvPip_WithIndexArgs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	routine.InitPool(64)
	record := path.Join(t.TempDir(), "uv_args.log")
	fakeUv := writeFakeUv(t, record)

	pluginSourceDir := path.Join("testdata", "plugin-with-requirements")
	dec, err := decoder.NewFSPluginDecoder(pluginSourceDir)
	require.NoError(t, err)

	appCfg := &app.Config{UvPath: fakeUv, PythonEnvInitTimeout: 60, PipMirrorUrl: "https://pypi.org/simple", PipExtraIndexUrl: "https://x/simple, https://y/simple"}
	runtime, err := ConstructPluginRuntime(appCfg, dec)
	require.NoError(t, err)
	require.NoError(t, copyDir(pluginSourceDir, runtime.State.WorkingPath))
	_, err = runtime.createVirtualEnvironment(fakeUv)
	require.NoError(t, err)
	depType, err := runtime.detectDependencyFileType()
	require.NoError(t, err)
	require.NoError(t, runtime.installDependencies(fakeUv, depType))

	b, err := os.ReadFile(record)
	require.NoError(t, err)
	out := string(b)
	require.Contains(t, out, "pip install -i https://pypi.org/simple")
	require.Contains(t, out, "--extra-index-url https://x/simple")
	require.Contains(t, out, "--extra-index-url https://y/simple")
	require.Contains(t, out, "-r requirements.txt")
}

func TestBuildDependencyInstallEnv(t *testing.T) {
	// Normalize PATH for deterministic assertion
	t.Setenv("PATH", "/bin:/usr/bin")
	r := &LocalPluginRuntime{appConfig: &app.Config{
		PipMirrorUrl:     "https://uv.example/simple",
		PipExtraIndexUrl: "https://b/simple https://a/simple",
		HttpProxy:        "http://proxy.internal:8080",
		HttpsProxy:       "https://secure-proxy.internal:8443",
		NoProxy:          "localhost,127.0.0.1",
	}}
	env := r.buildDependencyInstallEnv("/work/.venv")
	m := envSliceToMap(env)
	require.Equal(t, "/work/.venv", m["VIRTUAL_ENV"]) 
	require.Equal(t, "/bin:/usr/bin", m["PATH"]) 
	// hosts are sorted alphabetically in deriveTrustedHosts
	require.Equal(t, "a b uv.example", m["PIP_TRUSTED_HOST"]) 
	require.Equal(t, "http://proxy.internal:8080", m["HTTP_PROXY"]) 
	require.Equal(t, "https://secure-proxy.internal:8443", m["HTTPS_PROXY"]) 
	require.Equal(t, "localhost,127.0.0.1", m["NO_PROXY"]) 
}

func TestBuildDependencyInstallEnv_NoTrustedHostsOrProxy(t *testing.T) {
	r := &LocalPluginRuntime{appConfig: &app.Config{}}
	env := r.buildDependencyInstallEnv("/venv")
	m := envSliceToMap(env)
	require.Equal(t, "/venv", m["VIRTUAL_ENV"]) 
	// PATH must always be present; value depends on test runner env, so just assert key exists
	_, ok := m["PATH"]
	require.True(t, ok)
	require.NotContains(t, m, "PIP_TRUSTED_HOST")
	require.NotContains(t, m, "HTTP_PROXY")
	require.NotContains(t, m, "HTTPS_PROXY")
	require.NotContains(t, m, "NO_PROXY")
}

func shouldRunRealUV() bool { return os.Getenv("RUN_REAL_UV_TESTS") == "1" }

func TestInstallDependencies_UvSync(t *testing.T) {
	if testing.Short() || !shouldRunRealUV() {
		t.Skip("Skipping real UV install test; set RUN_REAL_UV_TESTS=1 to enable")
	}
	routine.InitPool(256)
	uvPath := findUVPath(t)
	pythonPath := findPythonPath(t)

	pluginSourceDir := path.Join("testdata", "plugin-with-pyproject")
	dec, err := decoder.NewFSPluginDecoder(pluginSourceDir)
	require.NoError(t, err)

	appCfg := &app.Config{PythonInterpreterPath: pythonPath, UvPath: uvPath, PythonEnvInitTimeout: 180}
	runtime, err := ConstructPluginRuntime(appCfg, dec)
	require.NoError(t, err)

	require.NoError(t, copyDir(pluginSourceDir, runtime.State.WorkingPath))

	venv, err := runtime.createVirtualEnvironment(uvPath)
	require.NoError(t, err)
	require.NotNil(t, venv)

	depType, err := runtime.detectDependencyFileType()
	require.NoError(t, err)
	require.Equal(t, pyprojectTomlFile, depType)
	require.NoError(t, runtime.installDependencies(uvPath, depType))

	// verify dify_plugin is installed in site-packages
	venvRoot := path.Join(runtime.State.WorkingPath, ".venv")
	entries, err := os.ReadDir(path.Join(venvRoot, "lib"))
	require.NoError(t, err)
	found := false
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "python") {
			sp := path.Join(venvRoot, "lib", e.Name(), "site-packages")
			if _, err := os.Stat(path.Join(sp, "dify_plugin")); err == nil {
				found = true
				break
			}
			matches, _ := filepath.Glob(path.Join(sp, "dify_plugin-*.dist-info"))
			if len(matches) > 0 {
				found = true
				break
			}
		}
	}
	require.True(t, found, "dify_plugin not found in site-packages")
}

func TestInstallDependencies_UvPip(t *testing.T) {
	if testing.Short() || !shouldRunRealUV() {
		t.Skip("Skipping real UV install test; set RUN_REAL_UV_TESTS=1 to enable")
	}
	routine.InitPool(256)
	uvPath := findUVPath(t)
	pythonPath := findPythonPath(t)

	pluginSourceDir := path.Join("testdata", "plugin-with-requirements")
	dec, err := decoder.NewFSPluginDecoder(pluginSourceDir)
	require.NoError(t, err)

	appCfg := &app.Config{PythonInterpreterPath: pythonPath, UvPath: uvPath, PythonEnvInitTimeout: 180}
	runtime, err := ConstructPluginRuntime(appCfg, dec)
	require.NoError(t, err)

	require.NoError(t, copyDir(pluginSourceDir, runtime.State.WorkingPath))

	venv, err := runtime.createVirtualEnvironment(uvPath)
	require.NoError(t, err)
	require.NotNil(t, venv)

	depType, err := runtime.detectDependencyFileType()
	require.NoError(t, err)
	require.Equal(t, requirementsTxtFile, depType)
	require.NoError(t, runtime.installDependencies(uvPath, depType))

	// verify dify_plugin is installed in site-packages
	venvRoot := path.Join(runtime.State.WorkingPath, ".venv")
	entries, err := os.ReadDir(path.Join(venvRoot, "lib"))
	require.NoError(t, err)
	found := false
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "python") {
			sp := path.Join(venvRoot, "lib", e.Name(), "site-packages")
			if _, err := os.Stat(path.Join(sp, "dify_plugin")); err == nil {
				found = true
				break
			}
			matches, _ := filepath.Glob(path.Join(sp, "dify_plugin-*.dist-info"))
			if len(matches) > 0 {
				found = true
				break
			}
		}
	}
	require.True(t, found, "dify_plugin not found in site-packages")
}
