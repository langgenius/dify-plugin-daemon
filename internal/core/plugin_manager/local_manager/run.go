package local_manager

import (
	"errors"
	"fmt"
	"os/exec"
	"sync"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/local_manager/stdio_holder"
	"github.com/langgenius/dify-plugin-daemon/internal/process"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

func (r *LocalPluginRuntime) gc() {
	if r.io_identity != "" {
		stdio_holder.Remove(r.io_identity)
	}

	if r.w != nil {
		close(r.w)
		r.w = nil
	}
}

func (r *LocalPluginRuntime) init() {
	r.w = make(chan bool)
	r.State.Status = entities.PLUGIN_RUNTIME_STATUS_LAUNCHING
}

func (r *LocalPluginRuntime) StartPlugin() error {
	defer log.Info("plugin %s stopped", r.Config.Identity())

	r.init()
	// start plugin
	e := exec.Command("bash", r.Config.Execution.Launch)
	e.Dir = r.State.RelativePath
	process.WrapProcess(e)

	// get writer
	stdin, err := e.StdinPipe()
	if err != nil {
		r.State.Status = entities.PLUGIN_RUNTIME_STATUS_RESTARTING
		return fmt.Errorf("get stdin pipe failed: %s", err.Error())
	}
	defer stdin.Close()

	stdout, err := e.StdoutPipe()
	if err != nil {
		r.State.Status = entities.PLUGIN_RUNTIME_STATUS_RESTARTING
		return fmt.Errorf("get stdout pipe failed: %s", err.Error())
	}
	defer stdout.Close()

	stderr, err := e.StderrPipe()
	if err != nil {
		r.State.Status = entities.PLUGIN_RUNTIME_STATUS_RESTARTING
		return fmt.Errorf("get stderr pipe failed: %s", err.Error())
	}
	defer stderr.Close()

	if err := e.Start(); err != nil {
		r.State.Status = entities.PLUGIN_RUNTIME_STATUS_RESTARTING
		return err
	}

	// add to subprocess manager
	process.NewProcess(e)
	defer process.RemoveProcess(e)

	defer func() {
		// wait for plugin to exit
		err = e.Wait()
		if err != nil {
			r.State.Status = entities.PLUGIN_RUNTIME_STATUS_RESTARTING
			log.Error("plugin %s exited with error: %s", r.Config.Identity(), err.Error())
		}

		r.gc()
	}()
	defer e.Process.Kill()

	log.Info("plugin %s started", r.Config.Identity())

	// setup stdio
	stdio := stdio_holder.Put(r.Config.Identity(), stdin, stdout, stderr)
	r.io_identity = stdio.GetID()
	defer stdio.Stop()

	wg := sync.WaitGroup{}
	wg.Add(2)

	// listen to plugin stdout
	routine.Submit(func() {
		defer wg.Done()
		stdio.StartStdout()
	})

	// listen to plugin stderr
	routine.Submit(func() {
		defer wg.Done()
		stdio.StartStderr()
	})

	// wait for plugin to exit
	err = stdio.Wait()
	if err != nil {
		return err
	}

	wg.Wait()

	// plugin has exited
	r.State.Status = entities.PLUGIN_RUNTIME_STATUS_PENDING
	return nil
}

func (r *LocalPluginRuntime) Wait() (<-chan bool, error) {
	if r.w == nil {
		return nil, errors.New("plugin not started")
	}
	return r.w, nil
}
