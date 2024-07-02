package local_manager

import (
	"fmt"
	"os/exec"
	"sync"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/stdio_holder"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

func (r *LocalPluginRuntime) StartPlugin() error {
	r.State.Status = entities.PLUGIN_RUNTIME_STATUS_LAUNCHING
	defer func() {
		r.io_identity = ""
	}()
	defer log.Info("plugin %s stopped", r.Config.Identity())

	// start plugin
	e := exec.Command("bash", "launch.sh")
	e.Dir = r.State.RelativePath

	// get writer
	stdin, err := e.StdinPipe()
	if err != nil {
		r.State.Status = entities.PLUGIN_RUNTIME_STATUS_RESTARTING
		e.Process.Kill()
		return fmt.Errorf("get stdin pipe failed: %s", err.Error())
	}

	stdout, err := e.StdoutPipe()
	if err != nil {
		r.State.Status = entities.PLUGIN_RUNTIME_STATUS_RESTARTING
		e.Process.Kill()
		return fmt.Errorf("get stdout pipe failed: %s", err.Error())
	}

	stderr, err := e.StderrPipe()
	if err != nil {
		r.State.Status = entities.PLUGIN_RUNTIME_STATUS_RESTARTING
		e.Process.Kill()
		return fmt.Errorf("get stderr pipe failed: %s", err.Error())
	}

	if err := e.Start(); err != nil {
		r.State.Status = entities.PLUGIN_RUNTIME_STATUS_RESTARTING
		return err
	}

	log.Info("plugin %s started", r.Config.Identity())

	stdio := stdio_holder.PutStdio(stdin, stdout, stderr)

	wg := sync.WaitGroup{}
	wg.Add(1)

	// listen to plugin stdout
	routine.Submit(func() {
		defer wg.Done()
		stdio.StartStdout()
	})

	err = stdio.StartStderr()
	if err != nil {
		r.State.Status = entities.PLUGIN_RUNTIME_STATUS_RESTARTING
		e.Process.Kill()
		return err
	}

	// wait for plugin to exit
	err = e.Wait()
	if err != nil {
		r.State.Status = entities.PLUGIN_RUNTIME_STATUS_RESTARTING
		return err
	}

	wg.Wait()

	// plugin has exited
	r.State.Status = entities.PLUGIN_RUNTIME_STATUS_PENDING
	return nil
}
