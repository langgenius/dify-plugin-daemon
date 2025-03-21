package local_runtime

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	version "github.com/hashicorp/go-version"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

//go:embed patches/0.0.1b70.ai_model.py.patch
var pythonPatches []byte

func (p *LocalPluginRuntime) InitPythonEnvironment() error {
	// check if virtual environment exists
	if _, err := os.Stat(path.Join(p.State.WorkingPath, ".venv")); err == nil {
		// check if venv is valid, try to find .venv/dify/plugin.json
		if _, err := os.Stat(path.Join(p.State.WorkingPath, ".venv/dify/plugin.json")); err != nil {
			// remove the venv and rebuild it
			os.RemoveAll(path.Join(p.State.WorkingPath, ".venv"))
		} else {
			// setup python interpreter path
			pythonPath, err := filepath.Abs(path.Join(p.State.WorkingPath, ".venv/bin/python"))
			if err != nil {
				return fmt.Errorf("failed to find python: %s", err)
			}
			p.pythonInterpreterPath = pythonPath
			// PATCH:
			//  plugin sdk version less than 0.0.1b70 contains a memory leak bug
			//  to reach a better user experience, we will patch it here using a patched file
			// https://github.com/langgenius/dify-plugin-sdks/commit/161045b65f708d8ef0837da24440ab3872821b3b
			if err := p.patchPluginSdk(path.Join(p.State.WorkingPath, "requirements.txt")); err != nil {
				log.Error("failed to patch the plugin sdk: %s", err)
			}
			return nil
		}
	}

	// execute init command, create a virtual environment
	success := false

	// using `from uv._find_uv import find_uv_bin; print(find_uv_bin())` to find uv path
	cmd := exec.Command(p.defaultPythonInterpreterPath, "-c", "from uv._find_uv import find_uv_bin; print(find_uv_bin())")
	cmd.Dir = p.State.WorkingPath
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to find uv path: %s", err)
	}

	uvPath := strings.TrimSpace(string(output))

	cmd = exec.Command(uvPath, "venv", ".venv")
	cmd.Dir = p.State.WorkingPath
	b := bytes.NewBuffer(nil)
	cmd.Stdout = b
	cmd.Stderr = b
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create virtual environment: %s, output: %s", err, b.String())
	}
	defer func() {
		// if init failed, remove the .venv directory
		if !success {
			os.RemoveAll(path.Join(p.State.WorkingPath, ".venv"))
		} else {
			// create dify/plugin.json
			pluginJsonPath := path.Join(p.State.WorkingPath, ".venv/dify/plugin.json")
			os.MkdirAll(path.Dir(pluginJsonPath), 0755)
			os.WriteFile(pluginJsonPath, []byte(`{"timestamp":`+strconv.FormatInt(time.Now().Unix(), 10)+`}`), 0644)
		}
	}()

	pythonPath, err := filepath.Abs(path.Join(p.State.WorkingPath, ".venv/bin/python"))
	if err != nil {
		return fmt.Errorf("failed to find python: %s", err)
	}

	if _, err := os.Stat(pythonPath); err != nil {
		return fmt.Errorf("failed to find python: %s", err)
	}

	p.pythonInterpreterPath = pythonPath

	// try find requirements.txt
	requirementsPath := path.Join(p.State.WorkingPath, "requirements.txt")
	if _, err := os.Stat(requirementsPath); err != nil {
		return fmt.Errorf("failed to find requirements.txt: %s", err)
	}

	// install dependencies
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	args := []string{"install"}

	if p.pipMirrorUrl != "" {
		args = append(args, "-i", p.pipMirrorUrl)
	}

	args = append(args, "-r", "requirements.txt")

	if p.pipVerbose {
		args = append(args, "-vvv")
	}

	if p.pipExtraArgs != "" {
		args = append(args, strings.Split(p.pipExtraArgs, " ")...)
	}

	args = append([]string{"pip"}, args...)

	virtualEnvPath := path.Join(p.State.WorkingPath, ".venv")
	cmd = exec.CommandContext(ctx, uvPath, args...)
	cmd.Env = append(cmd.Env, "VIRTUAL_ENV="+virtualEnvPath, "PATH="+os.Getenv("PATH"))
	if p.HttpProxy != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("HTTP_PROXY=%s", p.HttpProxy))
	}
	if p.HttpsProxy != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("HTTPS_PROXY=%s", p.HttpsProxy))
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

	routine.Submit(map[string]string{
		"module":   "plugin_manager",
		"function": "InitPythonEnvironment",
	}, func() {
		defer wg.Done()
		// read stdout
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if err != nil {
				break
			}
			log.Info("installing %s - %s", p.Config.Identity(), string(buf[:n]))
			lastActiveAt = time.Now()
		}
	})

	routine.Submit(map[string]string{
		"module":   "plugin_manager",
		"function": "InitPythonEnvironment",
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

	routine.Submit(map[string]string{
		"module":   "plugin_manager",
		"function": "InitPythonEnvironment",
	}, func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
				break
			}

			if time.Since(lastActiveAt) > time.Duration(p.pythonEnvInitTimeout)*time.Second {
				cmd.Process.Kill()
				errMsg.WriteString(fmt.Sprintf("init process exited due to no activity for %d seconds", p.pythonEnvInitTimeout))
				break
			}
		}
	})

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to install dependencies: %s, output: %s", err, errMsg.String())
	}

	compileArgs := []string{"-m", "compileall"}
	if p.pythonCompileAllExtraArgs != "" {
		compileArgs = append(compileArgs, strings.Split(p.pythonCompileAllExtraArgs, " ")...)
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

	routine.Submit(map[string]string{
		"module":   "plugin_manager",
		"function": "InitPythonEnvironment",
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

	routine.Submit(map[string]string{
		"module":   "plugin_manager",
		"function": "InitPythonEnvironment",
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

	// PATCH:
	//  plugin sdk version less than 0.0.1b70 contains a memory leak bug
	//  to reach a better user experience, we will patch it here using a patched file
	// https://github.com/langgenius/dify-plugin-sdks/commit/161045b65f708d8ef0837da24440ab3872821b3b
	if err := p.patchPluginSdk(requirementsPath); err != nil {
		log.Error("failed to patch the plugin sdk: %s", err)
	}

	success = true

	return nil
}

func (p *LocalPluginRuntime) patchPluginSdk(requirementsPath string) error {
	// get the version of the plugin sdk
	requirements, err := os.ReadFile(requirementsPath)
	if err != nil {
		return fmt.Errorf("failed to read requirements.txt: %s", err)
	}

	pluginSdkVersion, err := p.getPluginSdkVersion(string(requirements))
	if err != nil {
		log.Error("failed to get the version of the plugin sdk: %s", err)
		return nil
	}

	pluginSdkVersionObj, err := version.NewVersion(pluginSdkVersion)
	if err != nil {
		log.Error("failed to create the version: %s", err)
		return nil
	}

	if pluginSdkVersionObj.LessThan(version.Must(version.NewVersion("0.0.1b70"))) {
		// get dify-plugin path
		command := exec.Command(p.pythonInterpreterPath, "-c", "import importlib.util;print(importlib.util.find_spec('dify_plugin').origin)")
		command.Dir = p.State.WorkingPath
		output, err := command.Output()
		if err != nil {
			return fmt.Errorf("failed to get the path of the plugin sdk: %s", err)
		}

		pluginSdkPath := path.Dir(strings.TrimSpace(string(output)))
		patchPath := path.Join(pluginSdkPath, "interfaces/model/ai_model.py")
		if _, err := os.Stat(patchPath); err != nil {
			return fmt.Errorf("failed to find the patch file: %s", err)
		}

		// apply the patch
		if _, err := os.Stat(patchPath); err != nil {
			return fmt.Errorf("failed to find the patch file: %s", err)
		}

		if err := os.WriteFile(patchPath, pythonPatches, 0644); err != nil {
			return fmt.Errorf("failed to write the patch file: %s", err)
		}
	}

	return nil
}

func (p *LocalPluginRuntime) getPluginSdkVersion(requirements string) (string, error) {
	// using regex to find the version of the plugin sdk
	re := regexp.MustCompile(`(?:dify[_-]plugin)(?:~=|==)([0-9.a-z]+)`)
	matches := re.FindStringSubmatch(requirements)
	if len(matches) < 2 {
		return "", fmt.Errorf("failed to find the version of the plugin sdk")
	}

	return matches[1], nil
}
