package local_runtime

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"time"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/constants"
	routinepkg "github.com/langgenius/dify-plugin-daemon/pkg/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
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
	r.WalkNotifiers(func(notifier PluginRuntimeNotifier) {
		notifier.OnInstanceStarting()
	})

	// get the command to start the plugin
	e, err := r.getInstanceCmd()
	if err != nil {
		r.WalkNotifiers(func(notifier PluginRuntimeNotifier) {
			notifier.OnInstanceLaunchFailed(nil, err)
		})
		return err
	}

	stdin, stdout, stderr, err := r.getInstanceStdio(e)
	if err != nil {
		r.WalkNotifiers(func(notifier PluginRuntimeNotifier) {
			notifier.OnInstanceLaunchFailed(nil, err)
		})
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
		r.WalkNotifiers(func(notifier PluginRuntimeNotifier) {
			notifier.OnInstanceLaunchFailed(nil, err)
		})
		return err
	}

	// setup stdio
	instance := newPluginInstance(r.Config.Identity(), e, stdin, stdout, stderr, r.appConfig)

	// setup lifecycle notifier
	launchNotifier := newNotifierLifecycleSignal([]func(){cleanupIOHolders})
	instance.AddNotifier(launchNotifier)

	launchChannel := make(chan bool)

	// setup launch notifier
	instance.AddNotifier(&PluginInstanceNotifierTemplate{
		// the first heartbeat will trigger this
		OnInstanceReadyImpl: func(pi *PluginInstance) {
			// notify plugin started
			r.WalkNotifiers(func(notifier PluginRuntimeNotifier) {
				notifier.OnInstanceReady(instance)
			})
			// mark the instance as started
			instance.started = true
			// setup instance
			r.instanceLocker.Lock()
			r.instances = append(r.instances, instance)
			r.instanceLocker.Unlock()

			close(launchChannel)
		},
		OnInstanceShutdownImpl: func(pi *PluginInstance) {
			// remove the instance from the list
			r.instanceLocker.Lock()
			r.instances = slices.DeleteFunc(r.instances, func(instance *PluginInstance) bool {
				return instance.instanceId == pi.instanceId
			})
			r.instanceLocker.Unlock()

			if !instance.started {
				// if the instance is not started, it means the plugin is not ready
				// so we need to notify the caller that the plugin is not ready
				r.WalkNotifiers(func(notifier PluginRuntimeNotifier) {
					notifier.OnInstanceLaunchFailed(
						instance,
						fmt.Errorf("plugin failed to start: %v", instance.Error()),
					)
				})
			}
		},
	})

	success := false
	defer func() {
		// if start NewInstance failed, close the pipes, avoid resource leak
		if !success {
			cleanupIOHolders()
			r.WalkNotifiers(func(notifier PluginRuntimeNotifier) {
				notifier.OnInstanceLaunchFailed(instance, err)
			})
		}
	}()

	// listen to plugin stdout
	routine.Submit(
		routinepkg.Labels{
			routinepkg.RoutineLabelKeyModule:      "plugin_manager",
			routinepkg.RoutineLabelRuntimeKeyType: "local",
			routinepkg.RoutineLabelKeyMethod:      "StartStdout",
		},
		instance.StartStdout,
	)

	// listen to plugin stderr
	routine.Submit(
		routinepkg.Labels{
			routinepkg.RoutineLabelKeyModule:      "plugin_manager",
			routinepkg.RoutineLabelRuntimeKeyType: "local",
			routinepkg.RoutineLabelKeyMethod:      "StartStderr",
		},
		instance.StartStderr,
	)

	// wait for first heartbeat
	timeout := time.NewTimer(MAX_HEARTBEAT_INTERVAL)
	defer timeout.Stop()

	select {
	case <-timeout.C:
		instance.Stop()
		return fmt.Errorf("failed to start plugin as no heartbeat received")
	case <-launchChannel:
		// nop
	}

	// monitor plugin
	routine.Submit(
		routinepkg.Labels{
			routinepkg.RoutineLabelKeyModule:      "plugin_manager",
			routinepkg.RoutineLabelRuntimeKeyType: "local",
			routinepkg.RoutineLabelKeyMethod:      "Monitor",
		},
		func() {
			instance.Monitor()
		},
	)

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
	instance.GracefulStop(time.Duration(r.appConfig.PluginMaxExecutionTimeout) * time.Second)
	return nil
}
