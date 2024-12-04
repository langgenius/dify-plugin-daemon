package local_manager

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

func (p *LocalPluginRuntime) InitPythonEnvironment() error {
	// check if virtual environment exists
	if _, err := os.Stat(path.Join(p.State.WorkingPath, ".venv")); err == nil {
		// setup python interpreter path
		pythonPath, err := filepath.Abs(path.Join(p.State.WorkingPath, ".venv/bin/python"))
		if err != nil {
			return fmt.Errorf("failed to find python: %s", err)
		}
		p.pythonInterpreterPath = pythonPath
		return nil
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

	cmd = exec.CommandContext(ctx, pipPath, "install", "-r", "requirements.txt")
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

	var err_msg strings.Builder
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
				err_msg.WriteString(string(buf[:n]))
				break
			} else if err == os.ErrClosed {
				break
			}

			if n > 0 {
				err_msg.WriteString(string(buf[:n]))
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

			if time.Since(lastActiveAt) > 60*time.Second {
				cmd.Process.Kill()
				err_msg.WriteString("init process exited due to long time no activity")
				break
			}
		}
	})

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to install dependencies: %s, output: %s", err, err_msg.String())
	}

	success = true
	return nil
}
