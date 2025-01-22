package local_runtime

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

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
			return nil
		}
	}

	// execute init command, create a virtual environment
	success := false

	cmd := exec.Command("bash", "-c", fmt.Sprintf("%s -m venv .venv", p.defaultPythonInterpreterPath))
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

	// try find python interpreter and pip
	pipPath, err := filepath.Abs(path.Join(p.State.WorkingPath, ".venv/bin/pip"))
	if err != nil {
		return fmt.Errorf("failed to find pip: %s", err)
	}

	pythonPath, err := filepath.Abs(path.Join(p.State.WorkingPath, ".venv/bin/python"))
	if err != nil {
		return fmt.Errorf("failed to find python: %s", err)
	}

	if _, err := os.Stat(pipPath); err != nil {
		return fmt.Errorf("failed to find pip: %s", err)
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

	if p.HttpProxy != "" {
		args = append(args, "--proxy", p.HttpProxy)
	} else if p.HttpsProxy != "" {
		args = append(args, "--proxy", p.HttpsProxy)
	}

	if p.pipMirrorUrl != "" {
		args = append(args, "-i", p.pipMirrorUrl)
	}

	args = append(args, "-r", "requirements.txt")

	cmd = exec.CommandContext(ctx, pipPath, args...)
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

	success = true
	return nil
}
