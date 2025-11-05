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
		notifier.OnRuntimeClose()
	})

	for atomic.LoadInt32(&r.scheduleStatus) == ScheduleStatusRunning {
		<-ticker.C

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
	}
}

func (r *LocalPluginRuntime) stopSchedule() {
	// set schedule status to stopped
	atomic.CompareAndSwapInt32(&r.scheduleStatus, ScheduleStatusRunning, ScheduleStatusStopped)
}

// Stop schedule loop, blocks until all instances were shutdown
func (r *LocalPluginRuntime) Stop() error {
	// inherit from PluginRuntime
	r.PluginRuntime.Stop()

	r.stopSchedule()

	// TODO: send stop signal and wait for all instances to be shutdown
	return nil
}

// GracefulStop stops the runtime gracefully
// Wait until all instances were gracefully shutdown and all sessions were closed
func (r *LocalPluginRuntime) GracefulStop() error {
	// stop schedule loop
	r.stopSchedule()

	// TODO: wait for all instances to be shutdown

	return nil
}
