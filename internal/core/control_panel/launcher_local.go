package controlpanel

import (
	"errors"
	"fmt"
	"sync"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/local_runtime"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

type LocalPluginInstanceLifetime struct {
	onStarting        func(instance *local_runtime.PluginInstance)
	onReady           func(instance *local_runtime.PluginInstance)
	onFailed          func(instance *local_runtime.PluginInstance, err error)
	onShutdown        func(instance *local_runtime.PluginInstance)
	onScaleDownFailed func(err error)
	onRuntimeClose    func()
}

func (c *LocalPluginInstanceLifetime) OnInstanceStarting(instance *local_runtime.PluginInstance) {
	if c.onStarting != nil {
		c.onStarting(instance)
	}
}

func (c *LocalPluginInstanceLifetime) OnInstanceReady(instance *local_runtime.PluginInstance) {
	if c.onReady != nil {
		c.onReady(instance)
	}
}

func (c *LocalPluginInstanceLifetime) OnInstanceLaunchFailed(instance *local_runtime.PluginInstance, err error) {
	if c.onFailed != nil {
		c.onFailed(instance, err)
	}
}

func (c *LocalPluginInstanceLifetime) OnInstanceShutdown(instance *local_runtime.PluginInstance) {
	if c.onShutdown != nil {
		c.onShutdown(instance)
	}
}

func (c *LocalPluginInstanceLifetime) OnInstanceScaleDownFailed(err error) {
	if c.onScaleDownFailed != nil {
		c.onScaleDownFailed(err)
	}
}

func (c *LocalPluginInstanceLifetime) OnRuntimeClose() {
	if c.onRuntimeClose != nil {
		c.onRuntimeClose()
	}
}

// Launches a local plugin runtime
//
// Returns a channel that notifies if the process finished (both success and failed)
func (c *ControlPanel) LaunchLocalPlugin(
	uniquePluginIdentifier plugin_entities.PluginUniqueIdentifier,
) (*local_runtime.LocalPluginRuntime, <-chan error, error) {
	// acquire semaphore, this semaphore will be released
	c.localPluginLaunchingSemaphore <- true

	// notify new runtime is starting
	c.WalkNotifiers(func(notifier ControlPanelNotifier) {
		notifier.OnLocalRuntimeStarting(uniquePluginIdentifier)
	})

	// launch and wait for ready or error
	runtime, err := c.getLocalPluginRuntime(uniquePluginIdentifier)
	if err != nil {
		err = errors.Join(err, fmt.Errorf("failed to get local plugin runtime"))
		// notify new runtime launch failed
		c.WalkNotifiers(func(notifier ControlPanelNotifier) {
			notifier.OnLocalRuntimeStartFailed(uniquePluginIdentifier, err)
		})
		// release semaphore
		<-c.localPluginLaunchingSemaphore
		return nil, nil, err
	}

	// init environment
	if err := runtime.InitEnvironment(); err != nil {
		err = errors.Join(err, fmt.Errorf("failed to init environment"))
		// notify new runtime launch failed
		c.WalkNotifiers(func(notifier ControlPanelNotifier) {
			notifier.OnLocalRuntimeStartFailed(uniquePluginIdentifier, err)
		})
		// release semaphore
		<-c.localPluginLaunchingSemaphore
		return nil, nil, err
	}

	once := sync.Once{}

	ch := make(chan error, 1)

	// mount a notifier to runtime
	lifetime := &LocalPluginInstanceLifetime{
		// only first instance ready will trigger this
		onReady: func(instance *local_runtime.PluginInstance) {
			once.Do(func() {
				// release semaphore
				<-c.localPluginLaunchingSemaphore
				// store the runtime
				c.localPluginRuntimes.Store(uniquePluginIdentifier, runtime)
				// notify new runtime ready
				c.WalkNotifiers(func(notifier ControlPanelNotifier) {
					notifier.OnLocalRuntimeReady(runtime)
				})
				ch <- nil
			})
		},
		// only first instance failed will trigger this
		onFailed: func(instance *local_runtime.PluginInstance, err error) {
			once.Do(func() {
				// release semaphore
				<-c.localPluginLaunchingSemaphore
				// notify new runtime launch failed
				c.WalkNotifiers(func(notifier ControlPanelNotifier) {
					notifier.OnLocalRuntimeStartFailed(uniquePluginIdentifier, err)
				})
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
		<-c.localPluginLaunchingSemaphore
		return nil, nil, err
	}

	return runtime, ch, nil
}
