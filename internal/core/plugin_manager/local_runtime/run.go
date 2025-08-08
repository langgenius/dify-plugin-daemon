// Package local_runtime handles the local plugin runtime management
package local_runtime

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/constants"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

// gc performs garbage collection for the LocalPluginRuntime
func (r *LocalPluginRuntime) gc() {
	if r.waitChan != nil {
		close(r.waitChan)
		r.waitChan = nil
	}
}

// Type returns the runtime type of the plugin
func (r *LocalPluginRuntime) Type() plugin_entities.PluginRuntimeType {
	return plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL
}

// getCmd prepares the exec.Cmd for the plugin based on its language
func (r *LocalPluginRuntime) getCmd() (*exec.Cmd, error) {
	if r.Config.Meta.Runner.Language == constants.Python {
		cmd := exec.Command(r.pythonInterpreterPath, "-m", r.Config.Meta.Runner.Entrypoint)
		cmd.Dir = r.State.WorkingPath
		cmd.Env = cmd.Environ()
		if r.HttpsProxy != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("HTTPS_PROXY=%s", r.HttpsProxy))
		}
		if r.HttpProxy != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("HTTP_PROXY=%s", r.HttpProxy))
		}
		if r.NoProxy != "" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("NO_PROXY=%s", r.NoProxy))
		}
		return cmd, nil
	}

	return nil, fmt.Errorf("unsupported language: %s", r.Config.Meta.Runner.Language)
}

// StartPlugin starts the plugin and manages its lifecycle
func (r *LocalPluginRuntime) StartPlugin() error {
	defer log.Info("plugin %s stopped", r.Config.Identity())
	defer func() {
		r.waitChanLock.Lock()
		for _, c := range r.waitStoppedChan {
			select {
			case c <- true:
			default:
			}
		}
		r.waitChanLock.Unlock()
	}()

	if r.isNotFirstStart {
		r.SetRestarting()
	} else {
		r.SetLaunching()
		r.isNotFirstStart = true
	}

	// reset wait chan
	r.waitChan = make(chan bool)
	// reset wait launched chan

	// start plugin
	e, err := r.getCmd()
	if err != nil {
		return err
	}

	e.Dir = r.State.WorkingPath
	// add env INSTALL_METHOD=local
	e.Env = append(e.Environ(), "INSTALL_METHOD=local", "PATH="+os.Getenv("PATH"))

	// get writer
	stdin, err := e.StdinPipe()
	if err != nil {
		return fmt.Errorf("get stdin pipe failed: %s", err.Error())
	}
	defer stdin.Close()

	// get stdout
	stdout, err := e.StdoutPipe()
	if err != nil {
		return fmt.Errorf("get stdout pipe failed: %s", err.Error())
	}
	defer stdout.Close()

	// get stderr
	stderr, err := e.StderrPipe()
	if err != nil {
		return fmt.Errorf("get stderr pipe failed: %s", err.Error())
	}
	defer stderr.Close()

	if err := e.Start(); err != nil {
		return fmt.Errorf("start plugin failed: %s", err.Error())
	}

	// setup stdio
	r.stdioHolder = newStdioHolder(r.Config.Identity(), stdin, stdout, stderr, &StdioHolderConfig{
		StdoutBufferSize:    r.stdoutBufferSize,
		StdoutMaxBufferSize: r.stdoutMaxBufferSize,
	})
	defer r.stdioHolder.Stop()

	defer func() {
		// wait for plugin to exit
		originalErr := e.Wait()
		if originalErr != nil {
			// get stdio
			var err error
			if r.stdioHolder != nil {
				stdioErr := r.stdioHolder.Error()
				if stdioErr != nil {
					err = errors.Join(originalErr, stdioErr)
				} else {
					err = originalErr
				}
			} else {
				err = originalErr
			}
			if err != nil {
				log.Error("plugin %s exited with error: %s", r.Config.Identity(), err.Error())
			} else {
				log.Error("plugin %s exited with unknown error", r.Config.Identity())
			}
		}

		r.gc()
	}()

	// ensure the plugin process is killed after the plugin exits
	defer e.Process.Kill()

	log.Info("plugin %s started", r.Config.Identity())

	wg := sync.WaitGroup{}
	wg.Add(2)

	// listen to plugin stdout
	routine.Submit(map[string]string{
		"module":   "plugin_manager",
		"type":     "local",
		"function": "StartStdout",
	}, func() {
		defer wg.Done()
		r.stdioHolder.StartStdout(func() {})
	})

	// listen to plugin stderr
	routine.Submit(map[string]string{
		"module":   "plugin_manager",
		"type":     "local",
		"function": "StartStderr",
	}, func() {
		defer wg.Done()
		r.stdioHolder.StartStderr()
	})

	// send started event
	r.waitChanLock.Lock()
	for _, c := range r.waitStartedChan {
		select {
		case c <- true:
		default:
		}
	}
	r.waitChanLock.Unlock()

	// wait for plugin to exit
	err = r.stdioHolder.Wait()
	if err != nil {
		return errors.Join(err, r.stdioHolder.Error())
	}
	wg.Wait()

	// plugin has exited
	return nil
}

// Wait returns a channel that will be closed when the plugin stops
func (r *LocalPluginRuntime) Wait() (<-chan bool, error) {
	if r.waitChan == nil {
		return nil, errors.New("plugin not started")
	}
	return r.waitChan, nil
}

// WaitStarted returns a channel that will receive true when the plugin starts
func (r *LocalPluginRuntime) WaitStarted() <-chan bool {
	c := make(chan bool)
	r.waitChanLock.Lock()
	r.waitStartedChan = append(r.waitStartedChan, c)
	r.waitChanLock.Unlock()
	return c
}

// WaitStopped returns a channel that will receive true when the plugin stops
func (r *LocalPluginRuntime) WaitStopped() <-chan bool {
	c := make(chan bool)
	r.waitChanLock.Lock()
	r.waitStoppedChan = append(r.waitStoppedChan, c)
	r.waitChanLock.Unlock()
	return c
}

// Stop stops the plugin
func (r *LocalPluginRuntime) Stop() {
	// inherit from PluginRuntime
	r.PluginRuntime.Stop()

	// get stdio
	if r.stdioHolder != nil {
		r.stdioHolder.Stop()
	}
}
