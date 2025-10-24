package local_runtime

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/constants"
)

// getCmd prepares the exec.Cmd for the plugin based on its language
func (r *LocalPluginRuntime) getInstanceCmd() (*exec.Cmd, error) {
	var cmd *exec.Cmd

	switch r.Config.Meta.Runner.Language {
	case constants.Python:
		pythonPath, err := r.getVirtualEnvironmentPythonPath()
		if err != nil {
			return nil, err
		}
		cmd = exec.Command(pythonPath, "-m", r.Config.Meta.Runner.Entrypoint)

	default:
		return nil, fmt.Errorf("unsupported language: %s", r.Config.Meta.Runner.Language)
	}

	cmd.Env = cmd.Environ()
	if r.appConfig.HttpsProxy != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("HTTPS_PROXY=%s", r.appConfig.HttpsProxy))
	}
	if r.appConfig.HttpProxy != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("HTTP_PROXY=%s", r.appConfig.HttpProxy))
	}
	if r.appConfig.NoProxy != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("NO_PROXY=%s", r.appConfig.NoProxy))
	}
	cmd.Env = append(cmd.Env, "INSTALL_METHOD=local", "PATH="+os.Getenv("PATH"))
	cmd.Dir = r.State.WorkingPath
	return cmd, nil
}

// getInstanceStdio gets the stdin, stdout, and stderr pipes for the plugin instance
// NOTE: close them after use
func (r *LocalPluginRuntime) getInstanceStdio(
	cmd *exec.Cmd,
) (io.WriteCloser, io.ReadCloser, io.ReadCloser, error) {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, nil, errors.Join(err, fmt.Errorf("get stdin pipe failed"))
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, nil, nil, errors.Join(err, fmt.Errorf("get stdout pipe failed"))
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdout.Close()
		stdin.Close()
		return nil, nil, nil, errors.Join(err, fmt.Errorf("get stderr pipe failed"))
	}

	return stdin, stdout, stderr, nil
}

// startNewInstance starts a new plugin instance
func (r *LocalPluginRuntime) startNewInstance() error {
	// get the command to start the plugin
	e, err := r.getInstanceCmd()
	if err != nil {
		return err
	}

	stdin, stdout, stderr, err := r.getInstanceStdio(e)
	if err != nil {
		return err
	}

	// cleanup IO holders
	cleanupIOHolders := func() {
		stdin.Close()
		stdout.Close()
		stderr.Close()
	}

	// start plugin process,
	if err := e.Start(); err != nil {
		cleanupIOHolders()
		return err
	}

	// setup stdio
	instance := newPluginInstance(r.Config.Identity(), e, stdin, stdout, stderr, r.appConfig)

	// setup launch notifier
	launchNotifier := newNotifierLifecycleSignal()
	instance.AddNotifier(launchNotifier)

	success := false
	defer func() {
		// if start NewInstance failed, close the pipes, avoid resource leak
		if !success {
			cleanupIOHolders()
		} else {
			// add cleanup callbacks to collect resources
			instance.AddNotifier(&NotifierShutdown{
				callbacks: []func(){
					cleanupIOHolders,
				},
			})
		}
	}()

	// notify plugin starting
	instance.WalkNotifiers(func(notifier PluginInstanceNotifier) {
		notifier.OnInstanceStarting(instance)
	})

	// listen to plugin stdout
	routine.Submit(
		map[string]string{"module": "plugin_manager", "type": "local", "function": "StartStdout"},
		instance.StartStdout,
	)

	// listen to plugin stderr
	routine.Submit(
		map[string]string{"module": "plugin_manager", "type": "local", "function": "StartStderr"},
		instance.StartStderr,
	)

	// wait for first heartbeat
	timeout := time.NewTimer(MAX_HEARTBEAT_INTERVAL)
	defer timeout.Stop()

	select {
	case <-timeout.C:
		instance.Stop()
		return fmt.Errorf("failed to start plugin as no heartbeat received")
	case <-launchNotifier.WaitLaunchSignal():
		// nop
	}

	// monitor plugin
	routine.Submit(
		map[string]string{"module": "plugin_manager", "type": "local", "function": "Monitor"},
		func() {
			instance.Monitor()
		},
	)

	// setup instance
	r.instanceLocker.Lock()
	r.instances = append(r.instances, instance)
	r.instanceLocker.Unlock()

	// notify plugin started
	instance.WalkNotifiers(func(notifier PluginInstanceNotifier) {
		notifier.OnInstanceReady(instance)
	})

	success = true
	return nil
}

func (r *LocalPluginRuntime) gracefullyStopLowestLoadInstance() error {
	// get the instance with the lowest load
	instance, err := r.pickLowestLoadInstance()
	if err != nil {
		return err
	}

	// gracefully shutdown the instance
	return instance.GracefulStop(MAX_GRACEFUL_STOP_INTERVAL)
}
