package local_runtime

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	routinepkg "github.com/langgenius/dify-plugin-daemon/pkg/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
)

func (p *LocalPluginRuntime) prepareUV() (string, error) {
	if p.uvPath != "" {
		return p.uvPath, nil
	}

	// using `from uv._find_uv import find_uv_bin; print(find_uv_bin())` to find uv path
	cmd := exec.Command(p.defaultPythonInterpreterPath, "-c", "from uv._find_uv import find_uv_bin; print(find_uv_bin())")
	cmd.Dir = p.State.WorkingPath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to find uv path: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func (p *LocalPluginRuntime) preparePipArgs() []string {
	args := []string{"install"}

	// Determine index URL precedence for pip install:
	indexURL := p.appConfig.PipMirrorUrl
	// Extra index URLs (comma or space separated); fallback to UV extras
	extra := p.appConfig.PipExtraIndexUrl
	args = addIndexArgs(args, indexURL, extra)

	// Derive trusted-host from index/extra URLs
	for _, h := range deriveTrustedHosts(indexURL, extra) {
		args = append(args, "--trusted-host", h)
	}

	args = append(args, "-r", "requirements.txt")

	if p.appConfig.PipVerbose {
		args = append(args, "-vvv")
	}

	if p.appConfig.PipExtraArgs != "" {
		extraArgs := strings.Split(p.appConfig.PipExtraArgs, " ")
		args = append(args, extraArgs...)
	}

	args = append([]string{"pip"}, args...)

	return args
}

func (p *LocalPluginRuntime) prepareSyncArgs() []string {
	args := []string{"sync", "--no-dev"}

	// Determine index URL precedence for uv sync:
	indexURL := p.appConfig.PipMirrorUrl
	// Extra index URLs; fallback to pip extras
	extra := p.appConfig.PipExtraIndexUrl
	args = addIndexArgs(args, indexURL, extra)

	if p.appConfig.PipVerbose {
		args = append(args, "-v")
	}

	if p.appConfig.PipExtraArgs != "" {
		extraArgs := strings.Split(p.appConfig.PipExtraArgs, " ")
		args = append(args, extraArgs...)
	}

	return args
}

func (p *LocalPluginRuntime) detectDependencyFileType() (PythonDependencyFileType, error) {
	pyprojectPath := path.Join(p.State.WorkingPath, string(pyprojectTomlFile))
	requirementsPath := path.Join(p.State.WorkingPath, string(requirementsTxtFile))

	if _, err := os.Stat(pyprojectPath); err == nil {
		return pyprojectTomlFile, nil
	}

	if _, err := os.Stat(requirementsPath); err == nil {
		return requirementsTxtFile, nil
	}

	return "", fmt.Errorf("neither %s nor %s found in plugin directory", pyprojectTomlFile, requirementsTxtFile)
}

// buildDependencyInstallEnv builds environment variables for dependency installation.
func (p *LocalPluginRuntime) buildDependencyInstallEnv(virtualEnvPath string) []string {
	env := []string{
		"VIRTUAL_ENV=" + virtualEnvPath,
		"PATH=" + os.Getenv("PATH"),
	}

	// Provide PIP_TRUSTED_HOST (space-separated) for pip under uv
	pipIndex := p.appConfig.PipMirrorUrl
	pipExtra := p.appConfig.PipExtraIndexUrl
	if hosts := deriveTrustedHosts(pipIndex, pipExtra); len(hosts) > 0 {
		env = append(env, fmt.Sprintf("PIP_TRUSTED_HOST=%s", strings.Join(hosts, " ")))
	}

	if p.appConfig.HttpProxy != "" {
		env = append(env, fmt.Sprintf("HTTP_PROXY=%s", p.appConfig.HttpProxy))
	}
	if p.appConfig.HttpsProxy != "" {
		env = append(env, fmt.Sprintf("HTTPS_PROXY=%s", p.appConfig.HttpsProxy))
	}
	if p.appConfig.NoProxy != "" {
		env = append(env, fmt.Sprintf("NO_PROXY=%s", p.appConfig.NoProxy))
	}
	return env
}

// withOpLogging wraps fn with standardized start/finish logging and duration measurement.
func (p *LocalPluginRuntime) withOpLogging(op string, kvs []any, fn func() error) error {
	startAt := time.Now()
	base := []any{"plugin", p.Config.Identity()}
	log.Info("starting "+op, append(base, kvs...)...)
	err := fn()
	if err != nil {
		fields := append(append([]any{}, base...), kvs...)
		fields = append(fields, "duration", time.Since(startAt).String(), "error", err)
		log.Error(op+" failed", fields...)
		return err
	}
	fields := append(append([]any{}, base...), kvs...)
	fields = append(fields, "duration", time.Since(startAt).String())
	log.Info(op+" finished", fields...)
	return nil
}

func (p *LocalPluginRuntime) installDependencies(
	uvPath string,
	dependencyFileType PythonDependencyFileType,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	var args []string
	var methodLabel string
	switch dependencyFileType {
	case pyprojectTomlFile:
		args = p.prepareSyncArgs()
		methodLabel = "uv sync"
	case requirementsTxtFile:
		args = p.preparePipArgs()
		methodLabel = "uv pip install"
	default:
		return fmt.Errorf("unsupported dependency file type: %s", dependencyFileType)
	}

	virtualEnvPath := path.Join(p.State.WorkingPath, ".venv")
	sanitized := sanitizeArgs(args)

	return p.withOpLogging("dependency installation", []any{
		"method", methodLabel,
		"args", strings.Join(sanitized, " "),
	}, func() error {
		cmd := exec.CommandContext(ctx, uvPath, args...)
		cmd.Env = append(cmd.Env, p.buildDependencyInstallEnv(virtualEnvPath)...)
		cmd.Dir = p.State.WorkingPath

		// get stdout and stderr
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("failed to get stdout: %s", err)
		}
		defer stdout.Close()

		stderr, err := cmd.StderrPipe()
		if err != nil {
			return fmt.Errorf("failed to get stderr: %s", err)
		}
		defer stderr.Close()

		// start command
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start command: %s", err)
		}

		defer func() {
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
		}()

		var errMsg strings.Builder
		var errMu sync.Mutex
		var wg sync.WaitGroup
		wg.Add(2)

		var lastActiveAt atomic.Int64
		lastActiveAt.Store(time.Now().UnixNano())

		routine.Submit(routinepkg.Labels{
			routinepkg.RoutineLabelKeyModule: "plugin_manager",
			routinepkg.RoutineLabelKeyMethod: "InitPythonEnvironment",
		}, func() {
			defer wg.Done()
			// read stdout line by line
			scanner := bufio.NewScanner(stdout)
			buf := make([]byte, 0, 64*1024)
			scanner.Buffer(buf, 10*1024*1024)
			for scanner.Scan() {
				line := scanner.Text()
				log.Info("install deps", "plugin", p.Config.Identity(), "stream", "stdout", "line", line)
				lastActiveAt.Store(time.Now().UnixNano())
			}
			if err := scanner.Err(); err != nil {
				errMu.Lock()
				errMsg.WriteString("stdout scan error: ")
				errMsg.WriteString(err.Error())
				errMsg.WriteString("\n")
				errMu.Unlock()
				log.Warn("install deps", "plugin", p.Config.Identity(), "stream", "stdout", "scanner_err", err.Error())
			}
		})

		routine.Submit(routinepkg.Labels{
			routinepkg.RoutineLabelKeyModule: "plugin_manager",
			routinepkg.RoutineLabelKeyMethod: "InitPythonEnvironment",
		}, func() {
			defer wg.Done()
			// read stderr line by line
			scanner := bufio.NewScanner(stderr)
			buf := make([]byte, 0, 64*1024)
			scanner.Buffer(buf, 10*1024*1024)
			for scanner.Scan() {
				line := scanner.Text()
				errMu.Lock()
				errMsg.WriteString(line)
				errMsg.WriteString("\n")
				errMu.Unlock()
				log.Warn("install deps", "plugin", p.Config.Identity(), "stream", "stderr", "line", line)
				lastActiveAt.Store(time.Now().UnixNano())
			}
			if err := scanner.Err(); err != nil {
				errMu.Lock()
				errMsg.WriteString("stderr scan error: ")
				errMsg.WriteString(err.Error())
				errMsg.WriteString("\n")
				errMu.Unlock()
				log.Warn("install deps", "plugin", p.Config.Identity(), "stream", "stderr", "scanner_err", err.Error())
			}
		})

		routine.Submit(routinepkg.Labels{
			routinepkg.RoutineLabelKeyModule: "plugin_manager",
			routinepkg.RoutineLabelKeyMethod: "InitPythonEnvironment",
		}, func() {
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
					break
				}

				if time.Since(time.Unix(0, lastActiveAt.Load())) > time.Duration(
					p.appConfig.PythonEnvInitTimeout,
				)*time.Second {
					cmd.Process.Kill()
					errMu.Lock()
					errMsg.WriteString(fmt.Sprintf(
						"init process exited due to no activity for %d seconds",
						p.appConfig.PythonEnvInitTimeout,
					))
					errMu.Unlock()
					break
				}
			}
		})

		wg.Wait()

		if err := cmd.Wait(); err != nil {
			return fmt.Errorf("failed to install dependencies: %s, output: %s", err, errMsg.String())
		}
		return nil
	})
}

// sanitizeArgs redacts credentials in any URL-like arguments to avoid leaking secrets in logs.
func sanitizeArgs(args []string) []string {
	// Match https://user:pass@ and https://user@
	reWithPass := regexp.MustCompile(`(https?://)[^/@:]+:[^/@]+@`)
	reUserOnly := regexp.MustCompile(`(https?://)[^/@:]+@`)

	out := make([]string, len(args))
	for i, a := range args {
		s := reWithPass.ReplaceAllString(a, "${1}****:****@")
		s = reUserOnly.ReplaceAllString(s, "${1}****:****@")
		out[i] = s
	}
	return out
}

type PythonVirtualEnvironment struct {
	pythonInterpreterPath string
}

var (
	ErrVirtualEnvironmentNotFound = errors.New("virtual environment not found")
	ErrVirtualEnvironmentInvalid  = errors.New("virtual environment is invalid")
)

type PythonDependencyFileType string

const (
	pyprojectTomlFile   PythonDependencyFileType = "pyproject.toml"
	requirementsTxtFile PythonDependencyFileType = "requirements.txt"
)

const (
	envPath          = ".venv"
	envPythonPath    = envPath + "/bin/python"
	envValidFlagFile = envPath + "/dify/plugin.json"
)

func (p *LocalPluginRuntime) checkPythonVirtualEnvironment() (*PythonVirtualEnvironment, error) {
	if _, err := os.Stat(path.Join(p.State.WorkingPath, envPath)); err != nil {
		return nil, ErrVirtualEnvironmentNotFound
	}

	pythonPath, err := filepath.Abs(path.Join(p.State.WorkingPath, envPythonPath))
	if err != nil {
		return nil, fmt.Errorf("failed to find python: %s", err)
	}

	if _, err := os.Stat(pythonPath); err != nil {
		return nil, fmt.Errorf("failed to find python: %s", err)
	}

	// check if dify/plugin.json exists
	if _, err := os.Stat(path.Join(p.State.WorkingPath, envValidFlagFile)); err != nil {
		return nil, ErrVirtualEnvironmentInvalid
	}

	return &PythonVirtualEnvironment{
		pythonInterpreterPath: pythonPath,
	}, nil
}

func (p *LocalPluginRuntime) deleteVirtualEnvironment() error {
	// check if virtual environment exists
	venvDir := path.Join(p.State.WorkingPath, envPath)
	if _, err := os.Stat(venvDir); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	log.Warn("deleting existing Python virtual environment", "plugin", p.Config.Identity(), "path", venvDir)
	return os.RemoveAll(venvDir)
}

func (p *LocalPluginRuntime) createVirtualEnvironment(
	uvPath string,
) (*PythonVirtualEnvironment, error) {
	cmd := exec.Command(uvPath, "venv", envPath, "--python", "3.12")
	cmd.Dir = p.State.WorkingPath
	b := bytes.NewBuffer(nil)
	cmd.Stdout = b
	cmd.Stderr = b
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to create virtual environment: %s, output: %s", err, b.String())
	}

	pythonPath, err := filepath.Abs(path.Join(p.State.WorkingPath, envPythonPath))
	if err != nil {
		return nil, fmt.Errorf("failed to find python: %s", err)
	}

	if _, err := os.Stat(pythonPath); err != nil {
		return nil, fmt.Errorf("failed to find python: %s", err)
	}

	// try find pyproject.toml or requirements.txt
	dependencyFileType, err := p.detectDependencyFileType()
	if err != nil {
		return nil, fmt.Errorf("failed to find dependency file: %s", err)
	}

	log.Info("detected dependency file", "plugin", p.Config.Identity(), "file", dependencyFileType)

	return &PythonVirtualEnvironment{
		pythonInterpreterPath: pythonPath,
	}, nil
}

func (p *LocalPluginRuntime) getRequirementsPath() string {
	return path.Join(p.State.WorkingPath, string(requirementsTxtFile))
}

func (p *LocalPluginRuntime) getDependencyFilePath() (string, error) {
	dependencyFileType, err := p.detectDependencyFileType()
	if err != nil {
		return "", err
	}
	return path.Join(p.State.WorkingPath, string(dependencyFileType)), nil
}

func (p *LocalPluginRuntime) markVirtualEnvironmentAsValid() error {
	// pluginIdentityPath is a file that contains the timestamp of the virtual environment
	// which is used to mark the virtual environment as valid (All dependencies were installed)

	pluginJsonPath := path.Join(p.State.WorkingPath, envValidFlagFile)

	if err := os.MkdirAll(path.Dir(pluginJsonPath), 0755); err != nil {
		return fmt.Errorf("failed to create %s/dify directory: %s", envPath, err)
	}

	// write plugin.json
	if err := os.WriteFile(
		pluginJsonPath,
		[]byte(`{"timestamp":`+strconv.FormatInt(time.Now().Unix(), 10)+`}`),
		0644,
	); err != nil {
		return fmt.Errorf("failed to write plugin.json: %s", err)
	}

	return nil
}

// splitByCommaOrSpace splits a list like "a,b c" into tokens.
func splitByCommaOrSpace(s string) []string {
	// replace comma with space then split by spaces
	s = strings.ReplaceAll(s, ",", " ")
	fields := strings.Fields(s)
	return fields
}

// selectURL returns the first non-empty URL from the provided list.
func selectURL(urls ...string) string {
	for _, u := range urls {
		if u != "" {
			return u
		}
	}
	return ""
}

// addIndexArgs appends index and extra-index URL arguments to args.
func addIndexArgs(args []string, indexURL string, extraIndexURL string) []string {
	if indexURL != "" {
		args = append(args, "-i", indexURL)
	}
	if extraIndexURL != "" {
		for _, u := range splitByCommaOrSpace(extraIndexURL) {
			if u != "" {
				args = append(args, "--extra-index-url", u)
			}
		}
	}
	return args
}

// deriveTrustedHosts parses hostnames from index/extra URLs and returns a de-duplicated list.
func deriveTrustedHosts(indexURL string, extraIndexURL string) []string {
	set := map[string]struct{}{}
	add := func(raw string) {
		if strings.TrimSpace(raw) == "" {
			return
		}
		u, err := url.Parse(raw)
		if err != nil || u.Host == "" {
			return
		}
		host := u.Host
		if i := strings.Index(host, ":"); i >= 0 {
			host = host[:i]
		}
		set[host] = struct{}{}
	}
	add(indexURL)
	for _, raw := range splitByCommaOrSpace(extraIndexURL) {
		add(raw)
	}
	out := make([]string, 0, len(set))
	for h := range set {
		out = append(out, h)
	}
	// preserve deterministic order: sort hostnames
	sort.Strings(out)
	return out
}

func (p *LocalPluginRuntime) preCompile(
	pythonPath string,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	compileArgs := []string{"-m", "compileall"}
	if p.appConfig.PythonCompileAllExtraArgs != "" {
		compileArgs = append(compileArgs, strings.Split(p.appConfig.PythonCompileAllExtraArgs, " ")...)
	}
	compileArgs = append(compileArgs, ".")

	// pre-compile the plugin to avoid costly compilation on first invocation
	compileCmd := exec.CommandContext(ctx, pythonPath, compileArgs...)
	compileCmd.Dir = p.State.WorkingPath

	// get stdout and stderr
	compileStdout, err := compileCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout: %s", err)
	}
	defer compileStdout.Close()

	compileStderr, err := compileCmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr: %s", err)
	}
	defer compileStderr.Close()

	// start command
	if err := compileCmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %s", err)
	}
	defer func() {
		if compileCmd.Process != nil {
			compileCmd.Process.Kill()
		}
	}()

	var compileErrMsg strings.Builder
	var compileWg sync.WaitGroup
	compileWg.Add(2)

	routine.Submit(routinepkg.Labels{
		routinepkg.RoutineLabelKeyModule: "plugin_manager",
		routinepkg.RoutineLabelKeyMethod: "InitPythonEnvironment",
	}, func() {
		defer compileWg.Done()
		// read compileStdout
		for {
			buf := make([]byte, 102400)
			n, err := compileStdout.Read(buf)
			if err != nil {
				break
			}
			// split to first line
			lines := strings.Split(string(buf[:n]), "\n")

			for len(lines) > 0 && len(lines[0]) == 0 {
				lines = lines[1:]
			}

			if len(lines) > 0 {
				if len(lines) > 1 {
					log.Info("pre-compiling plugin", "plugin", p.Config.Identity(), "file", lines[0], "more", true)
				} else {
					log.Info("pre-compiling plugin", "plugin", p.Config.Identity(), "file", lines[0])
				}
			}
		}
	})

	routine.Submit(routinepkg.Labels{
		routinepkg.RoutineLabelKeyModule: "plugin_manager",
		routinepkg.RoutineLabelKeyMethod: "InitPythonEnvironment",
	}, func() {
		defer compileWg.Done()
		// read stderr
		buf := make([]byte, 1024)
		for {
			n, err := compileStderr.Read(buf)
			if err != nil {
				break
			}
			compileErrMsg.WriteString(string(buf[:n]))
		}
	})

	compileWg.Wait()
	if err := compileCmd.Wait(); err != nil {
		// skip the error if the plugin is not compiled
		// ISSUE: for some weird reasons, plugins may reference to a broken sdk but it works well itself
		// we need to skip it but log the messages
		// https://github.com/langgenius/dify/issues/16292
		log.Warn("failed to pre-compile the plugin", "error", compileErrMsg.String())
	}

	log.Info("pre-loaded the plugin", "plugin", p.Config.Identity())

	// import dify_plugin to speedup the first launching
	// ISSUE: it takes too long to setup all the deps, that's why we choose to preload it
	importCmd := exec.CommandContext(ctx, pythonPath, "-c", "import dify_plugin")
	importCmd.Dir = p.State.WorkingPath
	importCmd.Output()

	return nil
}

func (p *LocalPluginRuntime) getVirtualEnvironmentPythonPath() (string, error) {
	// get the absolute path of the python interpreter

	pythonPath, err := filepath.Abs(path.Join(p.State.WorkingPath, envPythonPath))
	if err != nil {
		return "", fmt.Errorf("failed to join python path: %s", err)
	}

	if _, err := os.Stat(pythonPath); err != nil {
		return "", ErrVirtualEnvironmentNotFound
	}

	return pythonPath, nil
}
