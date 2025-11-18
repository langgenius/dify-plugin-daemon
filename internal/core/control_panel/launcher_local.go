package controlpanel

import (
	"errors"
	"fmt"
	"sync"

	"github.com/langgenius/dify-plugin-daemon/internal/core/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	routinepkg "github.com/langgenius/dify-plugin-daemon/pkg/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
)

// Launches a local plugin runtime
// This method initializes environment (pypi, uv, dependencies, etc.) for a plugin
// and then starts the schedule loop
//
// NOTE: this method also triggers notifiers added to ControlPanel
// signal that a plugin starts successfully or failed
//
// Returns a channel that notifies if the process finished (both success and failed)
func (c *ControlPanel) LaunchLocalPlugin(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (*local_runtime.LocalPluginRuntime, <-chan error, error) {
	c.localPluginInstallationLock.Lock(pluginUniqueIdentifier.String())

	// check if the plugin is already installed
	if _, exists := c.localPluginRuntimes.Load(pluginUniqueIdentifier); exists {
		return nil, nil, ErrorPluginAlreadyLaunched
	}

	// acquire semaphore, this semaphore will be released
	c.localPluginLaunchingSemaphore <- true

	releaseLockAndSemaphore := func() {
		// this lock avoids the same plugin to be installed concurrently
		c.localPluginInstallationLock.Unlock(pluginUniqueIdentifier.String())

		// release semaphore to allow next plugin to be installed
		<-c.localPluginLaunchingSemaphore
	}

	// notify new runtime is starting
	c.WalkNotifiers(func(notifier ControlPanelNotifier) {
		notifier.OnLocalRuntimeStarting(pluginUniqueIdentifier)
	})

	// launch and wait for ready or error
	runtime, decoder, err := c.buildLocalPluginRuntime(pluginUniqueIdentifier)
	if err != nil {
		err = errors.Join(err, fmt.Errorf("failed to get local plugin runtime"))
		// notify new runtime launch failed
		c.WalkNotifiers(func(notifier ControlPanelNotifier) {
			notifier.OnLocalRuntimeStartFailed(pluginUniqueIdentifier, err)
		})
		// release semaphore
		releaseLockAndSemaphore()
		return nil, nil, err
	}

	// init environment
	// whatever it's a user request to launch a plugin or a new plugin was found
	// by watch dog, initialize environment is a must
	if err := runtime.InitEnvironment(decoder); err != nil {
		err = errors.Join(err, fmt.Errorf("failed to init environment"))
		// notify new runtime launch failed
		c.WalkNotifiers(func(notifier ControlPanelNotifier) {
			notifier.OnLocalRuntimeStartFailed(pluginUniqueIdentifier, err)
		})
		// release semaphore
		releaseLockAndSemaphore()
		return nil, nil, err
	}

	once := sync.Once{}
	ch := make(chan error, 1)

	// mount a notifier to runtime
	lifetime := &local_runtime.PluginRuntimeNotifierTemplate{
		// only first instance ready will trigger this
		OnInstanceReadyImpl: func(pi *local_runtime.PluginInstance) {
			// ideally, `once` is not needed here as `onReady` should only be triggered once
			once.Do(func() {
				// store the runtime
				c.localPluginRuntimes.Store(pluginUniqueIdentifier, runtime)
				// notify new runtime ready
				c.WalkNotifiers(func(notifier ControlPanelNotifier) {
					notifier.OnLocalRuntimeReady(runtime)
				})
				// release semaphore
				releaseLockAndSemaphore()
				ch <- nil
			})
		},
		OnInstanceScaleUpImpl: func(i int32) {
			c.WalkNotifiers(func(notifier ControlPanelNotifier) {
				notifier.OnLocalRuntimeScaleUp(runtime, i)
			})
		},
		OnInstanceScaleDownImpl: func(i int32) {
			c.WalkNotifiers(func(notifier ControlPanelNotifier) {
				notifier.OnLocalRuntimeScaleDown(runtime, i)
			})
		},
		// only first instance failed will trigger this
		OnInstanceLaunchFailedImpl: func(pi *local_runtime.PluginInstance, err error) {
			once.Do(func() {
				// notify new runtime launch failed
				c.WalkNotifiers(func(notifier ControlPanelNotifier) {
					notifier.OnLocalRuntimeStartFailed(pluginUniqueIdentifier, err)
				})
				// release semaphore
				releaseLockAndSemaphore()
				ch <- err
			})
		},
		OnRuntimeCloseImpl: func() {
			// notify the plugin totally stopped
			c.WalkNotifiers(func(notifier ControlPanelNotifier) {
				notifier.OnLocalRuntimeStopped(runtime)
			})
		},
		OnRuntimeStopScheduleImpl: func() {
			// delete the runtime from the map
			// Even if the runtime is not ready, deleting it still makes sense
			// once a plugin is stopping schedule, all new requests to it need to be rejected
			// so just remove it from map
			c.localPluginRuntimes.Delete(pluginUniqueIdentifier)
			// notify the plugin is stopping
			c.WalkNotifiers(func(notifier ControlPanelNotifier) {
				notifier.OnLocalRuntimeStop(runtime)
			})
		},
	}
	runtime.AddNotifier(lifetime)

	// scale up, ensure at least one instance is running
	runtime.ScaleUp()

	// start schedule
	// NOTE: it's a async method, releasing semaphore here is not a good idea
	// implemented inside `LocalPluginLaunchingSemaphore`
	if err := runtime.Schedule(); err != nil {
		err = errors.Join(err, fmt.Errorf("failed to schedule local plugin runtime"))
		// notify new runtime launch failed
		c.WalkNotifiers(func(notifier ControlPanelNotifier) {
			notifier.OnLocalRuntimeStartFailed(pluginUniqueIdentifier, err)
		})
		// release semaphore
		releaseLockAndSemaphore()
		return nil, nil, err
	}

	return runtime, ch, nil
}

// Trigger a signal to stop a local plugin runtime
// Force shutdown a local plugin runtime
// Returns a channel that notifies if the process finished (both success and failed)
// forcefully, whatever the runtime is handling or not
func (c *ControlPanel) ShutdownLocalPluginForcefully(
	uniquePluginIdentifier plugin_entities.PluginUniqueIdentifier,
) (<-chan error, error) {
	runtime, exists := c.localPluginRuntimes.Load(uniquePluginIdentifier)
	if !exists {
		return nil, ErrLocalPluginRuntimeNotFound
	}

	ch := make(chan error, 1)

	routine.Submit(routinepkg.Labels{
		routinepkg.RoutineLabelKeyModule: "controlpanel",
		routinepkg.RoutineLabelKeyMethod: "ShutdownLocalPluginForcefully",
	}, func() {
		runtime.Stop(false)

		// trigger that the runtime is shutdown
		close(ch)
	})

	return ch, nil
}

// Gracefully shutdown a local plugin runtime
// Returns a channel that notifies if the process finished (both success and failed)
// The channel will be closed if graceful shutdown is done
// this method will wait for all requests to be processed in each instance
// and then stop it
func (c *ControlPanel) ShutdownLocalPluginGracefully(
	uniquePluginIdentifier plugin_entities.PluginUniqueIdentifier,
) (<-chan error, error) {
	runtime, exists := c.localPluginRuntimes.Load(uniquePluginIdentifier)
	if !exists {
		return nil, ErrLocalPluginRuntimeNotFound
	}

	ch := make(chan error, 1)

	// wait for runtime to be shutdown in a goroutine
	routine.Submit(routinepkg.Labels{
		routinepkg.RoutineLabelKeyModule: "controlpanel",
		routinepkg.RoutineLabelKeyMethod: "ShutdownLocalPluginGracefully",
	}, func() {
		runtime.GracefulStop(false)

		// trigger that the runtime has shutdown
		close(ch)
	})

	return ch, nil
}
