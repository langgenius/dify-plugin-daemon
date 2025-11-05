package controlpanel

import (
	"errors"
	"fmt"
	"sync"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

type LocalPluginInstanceLifetimeCallback struct {
	onStarting        func(instance *local_runtime.PluginInstance)
	onReady           func(instance *local_runtime.PluginInstance)
	onFailed          func(instance *local_runtime.PluginInstance, err error)
	onShutdown        func(instance *local_runtime.PluginInstance)
	onScaleDownFailed func(err error)
	onRuntimeClose    func()
}

func (c *LocalPluginInstanceLifetimeCallback) OnInstanceStarting(instance *local_runtime.PluginInstance) {
	if c.onStarting != nil {
		c.onStarting(instance)
	}
}

func (c *LocalPluginInstanceLifetimeCallback) OnInstanceReady(instance *local_runtime.PluginInstance) {
	if c.onReady != nil {
		c.onReady(instance)
	}
}

func (c *LocalPluginInstanceLifetimeCallback) OnInstanceLaunchFailed(instance *local_runtime.PluginInstance, err error) {
	if c.onFailed != nil {
		c.onFailed(instance, err)
	}
}

func (c *LocalPluginInstanceLifetimeCallback) OnInstanceShutdown(instance *local_runtime.PluginInstance) {
	if c.onShutdown != nil {
		c.onShutdown(instance)
	}
}

func (c *LocalPluginInstanceLifetimeCallback) OnInstanceScaleDownFailed(err error) {
	if c.onScaleDownFailed != nil {
		c.onScaleDownFailed(err)
	}
}

func (c *LocalPluginInstanceLifetimeCallback) OnRuntimeClose() {
	if c.onRuntimeClose != nil {
		c.onRuntimeClose()
	}
}

// Launches a local plugin runtime
// This method initializes environment (pypi, uv, dependencies, etc.) for a plugin
// and then starts the schedule loop
//
// Returns a channel that notifies if the process finished (both success and failed)
func (c *ControlPanel) LaunchLocalPlugin(
	uniquePluginIdentifier plugin_entities.PluginUniqueIdentifier,
) (*local_runtime.LocalPluginRuntime, <-chan error, error) {
	c.localPluginInstallationLock.Lock(uniquePluginIdentifier.String())

	// check if the plugin is already installed
	if _, exists := c.localPluginRuntimes.Load(uniquePluginIdentifier); exists {
		return nil, nil, ErrorPluginAlreadyLaunched
	}

	// acquire semaphore, this semaphore will be released
	c.localPluginLaunchingSemaphore <- true

	releaseLockAndSemaphore := func() {
		// this lock avoids the same plugin to be installed concurrently
		c.localPluginInstallationLock.Unlock(uniquePluginIdentifier.String())

		// release semaphore to allow next plugin to be installed
		<-c.localPluginLaunchingSemaphore
	}

	// notify new runtime is starting
	c.WalkNotifiers(func(notifier ControlPanelNotifier) {
		notifier.OnLocalRuntimeStarting(uniquePluginIdentifier)
	})

	// launch and wait for ready or error
	runtime, decoder, err := c.buildLocalPluginRuntime(uniquePluginIdentifier)
	if err != nil {
		err = errors.Join(err, fmt.Errorf("failed to get local plugin runtime"))
		// notify new runtime launch failed
		c.WalkNotifiers(func(notifier ControlPanelNotifier) {
			notifier.OnLocalRuntimeStartFailed(uniquePluginIdentifier, err)
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
			notifier.OnLocalRuntimeStartFailed(uniquePluginIdentifier, err)
		})
		// release semaphore
		releaseLockAndSemaphore()
		return nil, nil, err
	}

	once := sync.Once{}
	ch := make(chan error, 1)

	// mount a notifier to runtime
	lifetime := &LocalPluginInstanceLifetimeCallback{
		// only first instance ready will trigger this
		onReady: func(instance *local_runtime.PluginInstance) {
			// ideally, `once` is not needed here as `onReady` should only be triggered once
			once.Do(func() {
				// store the runtime
				c.localPluginRuntimes.Store(uniquePluginIdentifier, runtime)
				// notify new runtime ready
				c.WalkNotifiers(func(notifier ControlPanelNotifier) {
					notifier.OnLocalRuntimeReady(runtime)
				})
				// release semaphore
				releaseLockAndSemaphore()
				ch <- nil
			})
		},
		// only first instance failed will trigger this
		onFailed: func(instance *local_runtime.PluginInstance, err error) {
			once.Do(func() {
				// notify new runtime launch failed
				c.WalkNotifiers(func(notifier ControlPanelNotifier) {
					notifier.OnLocalRuntimeStartFailed(uniquePluginIdentifier, err)
				})
				// release semaphore
				releaseLockAndSemaphore()
				ch <- err
			})
		},
		onRuntimeClose: func() {
			// delete the runtime from the map
			// Even if the runtime is not ready, deleting it still makes sense
			c.localPluginRuntimes.Delete(uniquePluginIdentifier)
		},
	}
	runtime.AddNotifier(lifetime)

	// start schedule
	// NOTE: it's a async method, releasing semaphore here is not a good idea
	// implemented inside `LocalPluginLaunchingSemaphore`
	if err := runtime.Schedule(); err != nil {
		err = errors.Join(err, fmt.Errorf("failed to schedule local plugin runtime"))
		// notify new runtime launch failed
		c.WalkNotifiers(func(notifier ControlPanelNotifier) {
			notifier.OnLocalRuntimeStartFailed(uniquePluginIdentifier, err)
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

	routine.Submit(map[string]string{
		"module": "controlpanel",
		"func":   "ShutdownLocalPluginForcefully",
	}, func() {
		err := runtime.Stop()
		if err != nil {
			ch <- err
		}

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
	routine.Submit(map[string]string{
		"module": "controlpanel",
		"func":   "ShutdownLocalPluginGracefully",
	}, func() {
		err := runtime.GracefulStop()
		if err != nil {
			ch <- err
		}

		// trigger that the runtime is shutdown
		close(ch)
	})

	return ch, nil
}
