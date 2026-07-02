package local_runtime

import (
	"slices"
	"sync/atomic"
	"time"

	routinepkg "github.com/langgenius/dify-plugin-daemon/pkg/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
)

const (
	ScheduleLoopInterval = 5 * time.Second
)

// Start schedule loop, it's a routine method will never block
func (r *LocalPluginRuntime) Schedule() error {
	if !atomic.CompareAndSwapInt32(&r.scheduleStatus, ScheduleStatusStopped, ScheduleStatusRunning) {
		// runtime already started
		return ErrRuntimeAlreadyStarted
	}

	// start schedule loop
	routine.Submit(routinepkg.Labels{
		routinepkg.RoutineLabelKeyModule: "local_runtime",
		routinepkg.RoutineLabelKeyMethod: "scheduleLoop",
	}, r.scheduleLoop)

	return nil
}

// Increase replicas
func (r *LocalPluginRuntime) ScaleUp() {
	atomic.AddInt32(&r.instanceNums, 1)
	r.WalkNotifiers(func(notifier PluginRuntimeNotifier) {
		notifier.OnInstanceScaleUp(r.instanceNums)
	})
}

// Decrease replicas
func (r *LocalPluginRuntime) ScaleDown() {
	atomic.AddInt32(&r.instanceNums, -1)
	r.WalkNotifiers(func(notifier PluginRuntimeNotifier) {
		notifier.OnInstanceScaleDown(r.instanceNums)
	})
}

func (r *LocalPluginRuntime) scheduleLoop() {
	// once it's not match, scale it
	ticker := time.NewTicker(ScheduleLoopInterval)
	defer ticker.Stop()

	for atomic.LoadInt32(&r.scheduleStatus) == ScheduleStatusRunning {
		// check if the instance nums is match
		r.instanceLocker.Lock()
		removed := r.pruneStoppedInstancesLocked()
		currentInstanceNums := len(r.instances)
		r.instanceLocker.Unlock()

		if removed > 0 {
			log.Warn(
				"pruned stopped local plugin instances",
				"plugin", r.Config.Identity(),
				"removed", removed,
				"live_instances", currentInstanceNums,
				"expected_instances", atomic.LoadInt32(&r.instanceNums),
			)
		}

		// if the current instance nums is less than the expected instance nums, start a new instance
		if currentInstanceNums < int(r.instanceNums) {
			// start a new instance
			if err := r.startNewInstance(); err != nil {
				// notify callers that a new instance failed to start
				r.WalkNotifiers(func(notifier PluginRuntimeNotifier) {
					notifier.OnInstanceLaunchFailed(nil, err)
				})
			} else {
				// notify callers that a new instance started
				r.WalkNotifiers(func(notifier PluginRuntimeNotifier) {
					notifier.OnInstanceStarting()
				})
			}
		} else if currentInstanceNums > int(r.instanceNums) {
			// gracefully shutdown the instance
			if err := r.gracefullyStopLowestLoadInstance(); err != nil {
				// notify callers that failed to gracefully stop a instance
				r.WalkNotifiers(func(notifier PluginRuntimeNotifier) {
					notifier.OnInstanceScaleDownFailed(err)
				})
			}
		}

		// wait for the next tick or an explicit kick after a dead instance is evicted.
		select {
		case <-ticker.C:
		case <-r.scheduleKick:
		}
	}

	// notify callers that the runtime is not running anymore
	r.WalkNotifiers(func(notifier PluginRuntimeNotifier) {
		notifier.OnRuntimeStopSchedule()
	})

	// wait for all instances to be shutdown
	r.waitForAllInstancesToBeShutdown()

	// notify callers that the runtime has been shutdown
	r.WalkNotifiers(func(notifier PluginRuntimeNotifier) {
		notifier.OnRuntimeClose()
	})
}

func (r *LocalPluginRuntime) nudgeSchedule() {
	if r.scheduleKick == nil {
		return
	}
	select {
	case r.scheduleKick <- struct{}{}:
	default:
	}
}

func (r *LocalPluginRuntime) pruneStoppedInstancesLocked() int {
	before := len(r.instances)
	r.instances = slices.DeleteFunc(r.instances, func(instance *PluginInstance) bool {
		return instance.IsStopped()
	})
	return before - len(r.instances)
}

func (r *LocalPluginRuntime) stopSchedule() {
	// set schedule status to stopped
	atomic.CompareAndSwapInt32(&r.scheduleStatus, ScheduleStatusRunning, ScheduleStatusStopped)
}

// Stop schedule loop, wait until all instances were shutdown
func (r *LocalPluginRuntime) Stop(async bool) {
	// inherit from PluginRuntime
	r.PluginRuntime.Stop()

	// stop schedule loop
	r.stopSchedule()

	// forcefully shutdown all instances
	r.forcefullyShutdownAllInstances()

	// wait for all instances to be shutdown
	if !async {
		r.waitForAllInstancesToBeShutdown()
	} else {
		routine.Submit(routinepkg.Labels{
			routinepkg.RoutineLabelKeyModule: "local_runtime",
			routinepkg.RoutineLabelKeyMethod: "waitForAllInstancesToBeShutdown",
		}, func() {
			r.waitForAllInstancesToBeShutdown()
		})
	}
}

// GracefulStop stops the runtime gracefully
// Wait until all instances were gracefully shutdown and all sessions were closed
func (r *LocalPluginRuntime) GracefulStop(async bool) {
	// inherit from PluginRuntime
	r.PluginRuntime.Stop()

	// stop schedule loop
	r.stopSchedule()

	// wait for all instances to be shutdown
	if !async {
		r.stopAndWaitForAllInstancesToBeShutdown()
	} else {
		routine.Submit(routinepkg.Labels{
			routinepkg.RoutineLabelKeyModule: "local_runtime",
			routinepkg.RoutineLabelKeyMethod: "stopAndWaitForAllInstancesToBeShutdown",
		}, func() {
			r.stopAndWaitForAllInstancesToBeShutdown()
		})
	}
}

// forcefully shutdown all instances, it's a async method which will not block
func (r *LocalPluginRuntime) forcefullyShutdownAllInstances() {
	for {
		r.instanceLocker.Lock()
		r.pruneStoppedInstancesLocked()
		instances := r.instances
		r.instanceLocker.Unlock()
		if len(instances) == 0 {
			break
		}
		instance := instances[0]
		instance.Stop()

		// sleep for 1 second to avoid busy waiting
		time.Sleep(time.Second * 1)
	}
}

// stop and wait for all instances to be shutdown
// please make sure to call this method after stop schedule loop
// otherwise new instances are going to start
func (r *LocalPluginRuntime) stopAndWaitForAllInstancesToBeShutdown() {
	for {
		r.instanceLocker.Lock()
		r.pruneStoppedInstancesLocked()
		instances := r.instances
		r.instanceLocker.Unlock()
		if len(instances) == 0 {
			break
		}
		instance := instances[0]
		instance.GracefulStop(time.Duration(r.appConfig.PluginMaxExecutionTimeout) * time.Second)

		// sleep for 1 second to avoid busy waiting
		time.Sleep(time.Second * 1)
	}
}

func (r *LocalPluginRuntime) waitForAllInstancesToBeShutdown() {
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	for {
		r.instanceLocker.Lock()
		r.pruneStoppedInstancesLocked()
		remaining := len(r.instances)
		r.instanceLocker.Unlock()
		if remaining == 0 {
			return
		}
		<-ticker.C
	}
}
