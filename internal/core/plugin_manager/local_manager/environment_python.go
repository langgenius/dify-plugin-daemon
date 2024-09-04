package local_manager

import (
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
	// execute init command, create a virtual environment
	success := false

	cmd := exec.Command("bash", "-c", "python3 -m venv .venv")
	cmd.Dir = p.State.WorkingPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create virtual environment: %s", err)
	}
	defer func() {
		// if init failed, remove the .venv directory
		if !success {
			os.RemoveAll(path.Join(p.State.WorkingPath, ".venv"))
		}
	}()

	// wait for the virtual environment to be created
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to create virtual environment: %s", err)
	}

	// try find python interpreter and pip
	pip_path, err := filepath.Abs(path.Join(p.State.WorkingPath, ".venv/bin/pip"))
	if err != nil {
		return fmt.Errorf("failed to find pip: %s", err)
	}

	python_path, err := filepath.Abs(path.Join(p.State.WorkingPath, ".venv/bin/python"))
	if err != nil {
		return fmt.Errorf("failed to find python: %s", err)
	}

	if _, err := os.Stat(pip_path); err != nil {
		return fmt.Errorf("failed to find pip: %s", err)
	}

	if _, err := os.Stat(python_path); err != nil {
		return fmt.Errorf("failed to find python: %s", err)
	}

	p.python_interpreter_path = python_path

	// try find requirements.txt
	requirements_path := path.Join(p.State.WorkingPath, "requirements.txt")
	if _, err := os.Stat(requirements_path); err != nil {
		return fmt.Errorf("failed to find requirements.txt: %s", err)
	}

	// install dependencies
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	cmd = exec.CommandContext(ctx, pip_path, "install", "-r", requirements_path)
	cmd.Dir = p.State.WorkingPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install dependencies: %s", err)
	}

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

	last_active_at := time.Now()

	routine.Submit(func() {
		defer wg.Done()
		// read stdout
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if err != nil {
				break
			}
			log.Info("installing %s - %s", p.Config.Identity(), string(buf[:n]))
			last_active_at = time.Now()
		}
	})

	routine.Submit(func() {
		defer wg.Done()
		// read stderr
		buf := make([]byte, 1024)
		for {
			n, err := stderr.Read(buf)
			if err != nil && err != os.ErrClosed {
				last_active_at = time.Now()
				err_msg.WriteString(string(buf[:n]))
				break
			} else if err == os.ErrClosed {
				break
			}

			if n > 0 {
				err_msg.WriteString(string(buf[:n]))
				last_active_at = time.Now()
			}
		}
	})

	routine.Submit(func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
				break
			}

			if time.Since(last_active_at) > 60*time.Second {
				cmd.Process.Kill()
				err_msg.WriteString("init process exited due to long time no activity")
				break
			}
		}
	})

	wg.Wait()

	if err_msg.Len() > 0 {
		return fmt.Errorf("install failed: %s", err_msg.String())
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to install dependencies: %s", err)
	}

	success = true
	return nil
}
