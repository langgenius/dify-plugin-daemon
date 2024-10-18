// Package local_manager handles the local plugin runtime management
package local_manager

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/langgenius/dify-plugin-daemon/internal/process"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/constants"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

// gc performs garbage collection for the LocalPluginRuntime
func (r *LocalPluginRuntime) gc() {
	if r.io_identity != "" {
		RemoveStdio(r.io_identity)
	}

	if r.wait_chan != nil {
		close(r.wait_chan)
		r.wait_chan = nil
	}
}

// init initializes the LocalPluginRuntime
func (r *LocalPluginRuntime) init() {
	// reset wait chan
	r.wait_chan = make(chan bool)
	// reset wait launched chan
	r.wait_launched_chan_once = sync.Once{}
	r.wait_launched_chan = make(chan error)

	r.SetLaunching()
}

// Type returns the runtime type of the plugin
func (r *LocalPluginRuntime) Type() plugin_entities.PluginRuntimeType {
	return plugin_entities.PLUGIN_RUNTIME_TYPE_LOCAL
}

// getCmd prepares the exec.Cmd for the plugin based on its language
func (r *LocalPluginRuntime) getCmd() (*exec.Cmd, error) {
	if r.Config.Meta.Runner.Language == constants.Python {
		cmd := exec.Command(r.python_interpreter_path, "-m", r.Config.Meta.Runner.Entrypoint)
		cmd.Dir = r.State.WorkingPath
		return cmd, nil
	}

	return nil, fmt.Errorf("unsupported language: %s", r.Config.Meta.Runner.Language)
}

// StartPlugin starts the plugin and manages its lifecycle
func (r *LocalPluginRuntime) StartPlugin() error {
	defer log.Info("plugin %s stopped", r.Config.Identity())
	defer func() {
		r.wait_chan_lock.Lock()
		for _, c := range r.wait_stopped_chan {
			select {
			case c <- true:
			default:
			}
		}
		r.wait_chan_lock.Unlock()
	}()

	r.init()

	// start plugin
	e, err := r.getCmd()
	if err != nil {
		r.wait_launched_chan_once.Do(func() {
			select {
			case r.wait_launched_chan <- err:
			default:
			}
			close(r.wait_launched_chan)
		})
		return err
	}

	e.Dir = r.State.WorkingPath
	// add env INSTALL_METHOD=local
	e.Env = append(e.Env, "INSTALL_METHOD=local", "PATH="+os.Getenv("PATH"))

	// NOTE: subprocess will be taken care of by subprocess manager
	// ensure all subprocess are killed when parent process exits, especially on Golang debugger
	process.WrapProcess(e)

	// notify launched, notify error if any
	notify_launched := func(err error) {
		r.wait_launched_chan_once.Do(func() {
			select {
			case r.wait_launched_chan <- err:
			default:
			}
			close(r.wait_launched_chan)
		})
	}

	// get writer
	stdin, err := e.StdinPipe()
	if err != nil {
		r.SetRestarting()
		err = fmt.Errorf("get stdin pipe failed: %s", err.Error())
		notify_launched(err)
		return err
	}
	defer stdin.Close()

	// get stdout
	stdout, err := e.StdoutPipe()
	if err != nil {
		r.SetRestarting()
		err = fmt.Errorf("get stdout pipe failed: %s", err.Error())
		notify_launched(err)
		return err
	}
	defer stdout.Close()

	// get stderr
	stderr, err := e.StderrPipe()
	if err != nil {
		r.SetRestarting()
		err = fmt.Errorf("get stderr pipe failed: %s", err.Error())
		notify_launched(err)
		return err
	}
	defer stderr.Close()

	if err := e.Start(); err != nil {
		r.SetRestarting()
		err = fmt.Errorf("start plugin failed: %s", err.Error())
		notify_launched(err)
		return err
	}

	// add to subprocess manager
	process.NewProcess(e)
	defer process.RemoveProcess(e)

	defer func() {
		// wait for plugin to exit
		err = e.Wait()
		if err != nil {
			r.SetRestarting()
			log.Error("plugin %s exited with error: %s", r.Config.Identity(), err.Error())
		}

		// close wait launched chan
		r.wait_launched_chan_once.Do(func() {
			close(r.wait_launched_chan)
		})

		r.gc()
	}()
	defer e.Process.Kill()

	log.Info("plugin %s started", r.Config.Identity())

	// setup stdio
	stdio := PutStdioIo(r.Config.Identity(), stdin, stdout, stderr)
	r.io_identity = stdio.GetID()
	defer stdio.Stop()

	wg := sync.WaitGroup{}
	wg.Add(2)

	// listen to plugin stdout
	routine.Submit(func() {
		defer wg.Done()
		stdio.StartStdout(func() {
			// get heartbeat, notify launched
			r.wait_launched_chan_once.Do(func() {
				close(r.wait_launched_chan)
			})
		})
	})

	// listen to plugin stderr
	routine.Submit(func() {
		defer wg.Done()
		stdio.StartStderr()
	})

	// send started event
	r.wait_chan_lock.Lock()
	for _, c := range r.wait_started_chan {
		select {
		case c <- true:
		default:
		}
	}
	r.wait_chan_lock.Unlock()

	// wait for plugin to exit
	err = stdio.Wait()
	if err != nil {
		return err
	}
	wg.Wait()

	// plugin has exited
	r.SetPending()
	return nil
}

// Wait returns a channel that will be closed when the plugin stops
func (r *LocalPluginRuntime) Wait() (<-chan bool, error) {
	if r.wait_chan == nil {
		return nil, errors.New("plugin not started")
	}
	return r.wait_chan, nil
}

// WaitStarted returns a channel that will receive true when the plugin starts
func (r *LocalPluginRuntime) WaitStarted() <-chan bool {
	c := make(chan bool)
	r.wait_chan_lock.Lock()
	r.wait_started_chan = append(r.wait_started_chan, c)
	r.wait_chan_lock.Unlock()
	return c
}

// WaitStopped returns a channel that will receive true when the plugin stops
func (r *LocalPluginRuntime) WaitStopped() <-chan bool {
	c := make(chan bool)
	r.wait_chan_lock.Lock()
	r.wait_stopped_chan = append(r.wait_stopped_chan, c)
	r.wait_chan_lock.Unlock()
	return c
}

// WaitLaunched returns a channel that will receive an error if the plugin fails to launch
func (r *LocalPluginRuntime) WaitLaunched() <-chan error {
	return r.wait_launched_chan
}
