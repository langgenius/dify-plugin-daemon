package local_runtime

import (
	"sync/atomic"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

const (
	ScheduleLoopInterval = 1 * time.Second
)

// Start schedule loop, it's a routine method will never block
func (r *LocalPluginRuntime) Schedule() error {
	if !atomic.CompareAndSwapInt32(&r.scheduleStatus, ScheduleStatusStopped, ScheduleStatusRunning) {
		// runtime already started
		return ErrRuntimeAlreadyStarted
	}

	// start schedule loop
	routine.Submit(map[string]string{
		"module": "local_runtime", "method": "scheduleLoop",
	}, r.scheduleLoop)

	return nil
}

func (r *LocalPluginRuntime) scheduleLoop() {
	// TODO: continuously check `instanceNums` and `instances`
	// once it's not match, scale it
	ticker := time.NewTicker(ScheduleLoopInterval)
	defer ticker.Stop()

	// notify callers that the runtime is not running anymore
	defer r.WalkNotifiers(func(notifier PluginRuntimeNotifier) {
		notifier.OnRuntimeStopSchedule()
	})

	for atomic.LoadInt32(&r.scheduleStatus) == ScheduleStatusRunning {
		// check if the instance nums is match
		r.instanceLocker.RLock()
		currentInstanceNums := len(r.instances)
		r.instanceLocker.RUnlock()

		// if the current instance nums is less than the expected instance nums, start a new instance
		if currentInstanceNums < r.instanceNums {
			// start a new instance
			if err := r.startNewInstance(); err != nil {
				// notify callers that a new instance failed to start
				r.WalkNotifiers(func(notifier PluginRuntimeNotifier) {
					notifier.OnInstanceLaunchFailed(nil, err)
				})
			}
		} else if currentInstanceNums > r.instanceNums {
			// gracefully shutdown the instance
			if err := r.gracefullyStopLowestLoadInstance(); err != nil {
				// notify callers that failed to gracefully stop a instance
				r.WalkNotifiers(func(notifier PluginRuntimeNotifier) {
					notifier.OnInstanceScaleDownFailed(err)
				})
			}
		}

		// wait for the next tick
		// this waiting must be done after all the schedule logic
		<-ticker.C
	}
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
		routine.Submit(map[string]string{
			"module": "local_runtime", "method": "waitForAllInstancesToBeShutdown",
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
		routine.Submit(map[string]string{
			"module": "local_runtime", "method": "stopAndWaitForAllInstancesToBeShutdown",
		}, func() {
			r.stopAndWaitForAllInstancesToBeShutdown()
		})
	}
}

// forcefully shutdown all instances, it's a async method which will not block
func (r *LocalPluginRuntime) forcefullyShutdownAllInstances() {
	instances := r.instances
	for _, instance := range instances {
		instance.Stop()
	}
}

// stop and wait for all instances to be shutdown
// please make sure to call this method after stop schedule loop
// otherwise new instances are going to start
func (r *LocalPluginRuntime) stopAndWaitForAllInstancesToBeShutdown() {
	instances := r.instances
	for _, instance := range instances {
		instance.GracefulStop(time.Duration(r.appConfig.PluginMaxExecutionTimeout) * time.Second)
	}
}

func (r *LocalPluginRuntime) waitForAllInstancesToBeShutdown() {
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	for len(r.instances) > 0 {
		<-ticker.C
	}

	// notify callers that the runtime is shutdown
	r.WalkNotifiers(func(notifier PluginRuntimeNotifier) {
		notifier.OnRuntimeClose()
	})
}
