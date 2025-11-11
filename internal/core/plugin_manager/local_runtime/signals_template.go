package local_runtime

type PluginRuntimeNotifierTemplate struct {
	PluginRuntimeNotifier

	OnInstanceStartingImpl        func(*PluginInstance)
	OnInstanceReadyImpl           func(*PluginInstance)
	OnInstanceFailedImpl          func(*PluginInstance, error)
	OnInstanceShutdownImpl        func(*PluginInstance)
	OnInstanceScaleDownFailedImpl func(error)
	OnRuntimeStopScheduleImpl     func()
	OnRuntimeCloseImpl            func()
}

func (t *PluginRuntimeNotifierTemplate) OnInstanceStarting(instance *PluginInstance) {
	if t.OnInstanceStartingImpl != nil {
		t.OnInstanceStartingImpl(instance)
	}
}

func (t *PluginRuntimeNotifierTemplate) OnInstanceReady(instance *PluginInstance) {
	if t.OnInstanceReadyImpl != nil {
		t.OnInstanceReadyImpl(instance)
	}
}

func (t *PluginRuntimeNotifierTemplate) OnInstanceFailed(instance *PluginInstance, err error) {
	if t.OnInstanceFailedImpl != nil {
		t.OnInstanceFailedImpl(instance, err)
	}
}

func (t *PluginRuntimeNotifierTemplate) OnInstanceShutdown(instance *PluginInstance) {
	if t.OnInstanceShutdownImpl != nil {
		t.OnInstanceShutdownImpl(instance)
	}
}

func (t *PluginRuntimeNotifierTemplate) OnInstanceScaleDownFailed(err error) {
	if t.OnInstanceScaleDownFailedImpl != nil {
		t.OnInstanceScaleDownFailedImpl(err)
	}
}

func (t *PluginRuntimeNotifierTemplate) OnRuntimeStopSchedule() {
	if t.OnRuntimeStopScheduleImpl != nil {
		t.OnRuntimeStopScheduleImpl()
	}
}

func (t *PluginRuntimeNotifierTemplate) OnRuntimeClose() {
	if t.OnRuntimeCloseImpl != nil {
		t.OnRuntimeCloseImpl()
	}
}
