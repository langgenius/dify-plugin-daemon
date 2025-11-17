package local_runtime

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
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

	if p.appConfig.PipMirrorUrl != "" {
		args = append(args, "-i", p.appConfig.PipMirrorUrl)
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

func (p *LocalPluginRuntime) installDependencies(
	uvPath string,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	args := p.preparePipArgs()

	virtualEnvPath := path.Join(p.State.WorkingPath, ".venv")
	cmd := exec.CommandContext(ctx, uvPath, args...)
	cmd.Env = append(cmd.Env, "VIRTUAL_ENV="+virtualEnvPath, "PATH="+os.Getenv("PATH"))
	if p.appConfig.HttpProxy != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("HTTP_PROXY=%s", p.appConfig.HttpProxy))
	}
	if p.appConfig.HttpsProxy != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("HTTPS_PROXY=%s", p.appConfig.HttpsProxy))
	}
	if p.appConfig.NoProxy != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("NO_PROXY=%s", p.appConfig.NoProxy))
	}
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
	var wg sync.WaitGroup
	wg.Add(2)

	lastActiveAt := time.Now()

	routine.Submit(routinepkg.Labels{
		routinepkg.RoutineLabelKeyModule: "plugin_manager",
		routinepkg.RoutineLabelKeyMethod: "InitPythonEnvironment",
	}, func() {
		defer wg.Done()
		// read stdout
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if err != nil {
				break
			}
			// FIXME: move the log to separated layer
			log.Info("installing %s - %s", p.Config.Identity(), string(buf[:n]))
			lastActiveAt = time.Now()
		}
	})

	routine.Submit(routinepkg.Labels{
		routinepkg.RoutineLabelKeyModule: "plugin_manager",
		routinepkg.RoutineLabelKeyMethod: "InitPythonEnvironment",
	}, func() {
		defer wg.Done()
		// read stderr
		buf := make([]byte, 1024)
		for {
			n, err := stderr.Read(buf)
			if err != nil && err != os.ErrClosed {
				lastActiveAt = time.Now()
				errMsg.WriteString(string(buf[:n]))
				break
			} else if err == os.ErrClosed {
				break
			}

			if n > 0 {
				errMsg.WriteString(string(buf[:n]))
				lastActiveAt = time.Now()
			}
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

			if time.Since(lastActiveAt) > time.Duration(
				p.appConfig.PythonEnvInitTimeout,
			)*time.Second {
				cmd.Process.Kill()
				errMsg.WriteString(fmt.Sprintf(
					"init process exited due to no activity for %d seconds",
					p.appConfig.PythonEnvInitTimeout,
				))
				break
			}
		}
	})

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to install dependencies: %s, output: %s", err, errMsg.String())
	}

	return nil
}

type PythonVirtualEnvironment struct {
	pythonInterpreterPath string
}

var (
	ErrVirtualEnvironmentNotFound = errors.New("virtual environment not found")
	ErrVirtualEnvironmentInvalid  = errors.New("virtual environment is invalid")
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
	if _, err := os.Stat(path.Join(p.State.WorkingPath, envPath)); err != nil {
		return nil
	}

	return os.RemoveAll(path.Join(p.State.WorkingPath, envPath))
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

	// try find requirements.txt
	requirementsPath := path.Join(p.State.WorkingPath, "requirements.txt")
	if _, err := os.Stat(requirementsPath); err != nil {
		return nil, fmt.Errorf("failed to find requirements.txt: %s", err)
	}

	return &PythonVirtualEnvironment{
		pythonInterpreterPath: pythonPath,
	}, nil
}

func (p *LocalPluginRuntime) getRequirementsPath() string {
	return path.Join(p.State.WorkingPath, "requirements.txt")
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
					log.Info("pre-compiling %s - %s...", p.Config.Identity(), lines[0])
				} else {
					log.Info("pre-compiling %s - %s", p.Config.Identity(), lines[0])
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
		log.Warn("failed to pre-compile the plugin: %s", compileErrMsg.String())
	}

	log.Info("pre-loaded the plugin %s", p.Config.Identity())

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
