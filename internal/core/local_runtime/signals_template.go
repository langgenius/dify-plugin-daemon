package local_runtime

import "github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"

type PluginRuntimeNotifierTemplate struct {
	OnInstanceStartingImpl        func()
	OnInstanceReadyImpl           func(*PluginInstance)
	OnInstanceLaunchFailedImpl    func(*PluginInstance, error)
	OnInstanceShutdownImpl        func(*PluginInstance)
	OnInstanceScaleUpImpl         func(int32)
	OnInstanceScaleDownImpl       func(int32)
	OnInstanceScaleDownFailedImpl func(error)
	OnInstanceLogImpl             func(*PluginInstance, plugin_entities.PluginLogEvent)
	OnRuntimeStopScheduleImpl     func()
	OnRuntimeCloseImpl            func()
}

func (t *PluginRuntimeNotifierTemplate) OnInstanceStarting() {
	if t.OnInstanceStartingImpl != nil {
		t.OnInstanceStartingImpl()
	}
}

func (t *PluginRuntimeNotifierTemplate) OnInstanceReady(instance *PluginInstance) {
	if t.OnInstanceReadyImpl != nil {
		t.OnInstanceReadyImpl(instance)
	}
}

func (t *PluginRuntimeNotifierTemplate) OnInstanceLaunchFailed(instance *PluginInstance, err error) {
	if t.OnInstanceLaunchFailedImpl != nil {
		t.OnInstanceLaunchFailedImpl(instance, err)
	}
}

func (t *PluginRuntimeNotifierTemplate) OnInstanceShutdown(instance *PluginInstance) {
	if t.OnInstanceShutdownImpl != nil {
		t.OnInstanceShutdownImpl(instance)
	}
}

func (t *PluginRuntimeNotifierTemplate) OnInstanceScaleUp(instanceNums int32) {
	if t.OnInstanceScaleUpImpl != nil {
		t.OnInstanceScaleUpImpl(instanceNums)
	}
}

func (t *PluginRuntimeNotifierTemplate) OnInstanceScaleDown(instanceNums int32) {
	if t.OnInstanceScaleDownImpl != nil {
		t.OnInstanceScaleDownImpl(instanceNums)
	}
}

func (t *PluginRuntimeNotifierTemplate) OnInstanceScaleDownFailed(err error) {
	if t.OnInstanceScaleDownFailedImpl != nil {
		t.OnInstanceScaleDownFailedImpl(err)
	}
}

func (t *PluginRuntimeNotifierTemplate) OnInstanceLog(
	instance *PluginInstance,
	event plugin_entities.PluginLogEvent,
) {
	if t.OnInstanceLogImpl != nil {
		t.OnInstanceLogImpl(instance, event)
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

type PluginInstanceNotifierTemplate struct {
	OnInstanceStartingImpl     func()
	OnInstanceReadyImpl        func(*PluginInstance)
	OnInstanceLaunchFailedImpl func(*PluginInstance, error)
	OnInstanceShutdownImpl     func(*PluginInstance)
	OnInstanceHeartbeatImpl    func(*PluginInstance)
	OnInstanceLogImpl          func(*PluginInstance, plugin_entities.PluginLogEvent)
	OnInstanceErrorLogImpl     func(*PluginInstance, error)
	OnInstanceWarningLogImpl   func(*PluginInstance, string)
	OnInstanceStdoutImpl       func(*PluginInstance, []byte)
	OnInstanceStderrImpl       func(*PluginInstance, []byte)
}

func (t *PluginInstanceNotifierTemplate) OnInstanceStarting() {
	if t.OnInstanceStartingImpl != nil {
		t.OnInstanceStartingImpl()
	}
}

func (t *PluginInstanceNotifierTemplate) OnInstanceReady(instance *PluginInstance) {
	if t.OnInstanceReadyImpl != nil {
		t.OnInstanceReadyImpl(instance)
	}
}

func (t *PluginInstanceNotifierTemplate) OnInstanceLaunchFailed(instance *PluginInstance, err error) {
	if t.OnInstanceLaunchFailedImpl != nil {
		t.OnInstanceLaunchFailedImpl(instance, err)
	}
}

func (t *PluginInstanceNotifierTemplate) OnInstanceShutdown(instance *PluginInstance) {
	if t.OnInstanceShutdownImpl != nil {
		t.OnInstanceShutdownImpl(instance)
	}
}

func (t *PluginInstanceNotifierTemplate) OnInstanceHeartbeat(instance *PluginInstance) {
	if t.OnInstanceHeartbeatImpl != nil {
		t.OnInstanceHeartbeatImpl(instance)
	}
}

func (t *PluginInstanceNotifierTemplate) OnInstanceLog(
	instance *PluginInstance,
	event plugin_entities.PluginLogEvent,
) {
	if t.OnInstanceLogImpl != nil {
		t.OnInstanceLogImpl(instance, event)
	}
}

func (t *PluginInstanceNotifierTemplate) OnInstanceErrorLog(instance *PluginInstance, err error) {
	if t.OnInstanceErrorLogImpl != nil {
		t.OnInstanceErrorLogImpl(instance, err)
	}
}

func (t *PluginInstanceNotifierTemplate) OnInstanceWarningLog(instance *PluginInstance, message string) {
	if t.OnInstanceWarningLogImpl != nil {
		t.OnInstanceWarningLogImpl(instance, message)
	}
}

func (t *PluginInstanceNotifierTemplate) OnInstanceStdout(instance *PluginInstance, data []byte) {
	if t.OnInstanceStdoutImpl != nil {
		t.OnInstanceStdoutImpl(instance, data)
	}
}

func (t *PluginInstanceNotifierTemplate) OnInstanceStderr(instance *PluginInstance, data []byte) {
	if t.OnInstanceStderrImpl != nil {
		t.OnInstanceStderrImpl(instance, data)
	}
}
