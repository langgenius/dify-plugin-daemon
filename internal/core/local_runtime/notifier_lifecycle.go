package local_runtime

import (
	"time"
)

type NotifierHeartbeat struct {
	afterShutdown []func()
}

func newNotifierLifecycleSignal(
	afterShutdown []func(),
) *NotifierHeartbeat {
	return &NotifierHeartbeat{
		afterShutdown: afterShutdown,
	}
}

func (n *NotifierHeartbeat) OnInstanceStarting() {

}

func (n *NotifierHeartbeat) OnInstanceReady(instance *PluginInstance) {

}

func (n *NotifierHeartbeat) OnInstanceLaunchFailed(instance *PluginInstance, err error) {

}

func (n *NotifierHeartbeat) OnInstanceShutdown(instance *PluginInstance) {
	instance.shutdown = true

	for _, callback := range n.afterShutdown {
		callback()
	}
}

func (n *NotifierHeartbeat) OnInstanceHeartbeat(instance *PluginInstance) {
	// update the last active time on each time the plugin sends data
	instance.lastActiveAt = time.Now()
}

func (n *NotifierHeartbeat) OnInstanceLog(instance *PluginInstance, message string) {

}

func (n *NotifierHeartbeat) OnInstanceErrorLog(instance *PluginInstance, err error) {

}

func (n *NotifierHeartbeat) OnInstanceWarningLog(instance *PluginInstance, message string) {

}

func (n *NotifierHeartbeat) OnInstanceStdout(instance *PluginInstance, data []byte) {
	instance.lastActiveAt = time.Now()
}

func (n *NotifierHeartbeat) OnInstanceStderr(instance *PluginInstance, data []byte) {
	instance.lastActiveAt = time.Now()
}
