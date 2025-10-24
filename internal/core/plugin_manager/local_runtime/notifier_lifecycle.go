package local_runtime

import (
	"sync"
	"time"
)

type NotifierHeartbeat struct {
	channel chan bool
	once    *sync.Once
}

func newNotifierLifecycleSignal() *NotifierHeartbeat {
	return &NotifierHeartbeat{
		channel: make(chan bool),
		once:    &sync.Once{},
	}
}

func (n *NotifierHeartbeat) OnInstanceStarting(instance *PluginInstance) {

}

func (n *NotifierHeartbeat) OnInstanceReady(instance *PluginInstance) {

}

func (n *NotifierHeartbeat) OnInstanceFailed(instance *PluginInstance, err error) {

}

func (n *NotifierHeartbeat) OnInstanceShutdown(instance *PluginInstance) {
	instance.shutdown = true
}

func (n *NotifierHeartbeat) OnInstanceHeartbeat(instance *PluginInstance) {
	// update the last active time on each time the plugin sends data
	instance.lastActiveAt = time.Now()

	n.once.Do(func() {
		// mark the instance as started
		instance.started = true
		close(n.channel)
	})
}

func (n *NotifierHeartbeat) OnInstanceLog(instance *PluginInstance, message string) {

}

func (n *NotifierHeartbeat) OnInstanceErrorLog(instance *PluginInstance, err error) {

}

func (n *NotifierHeartbeat) OnInstanceWarningLog(instance *PluginInstance, message string) {

}

func (n *NotifierHeartbeat) WaitLaunchSignal() <-chan bool {
	// wait for start signal
	return n.channel
}

func (n *NotifierHeartbeat) OnInstanceStdout(instance *PluginInstance, data []byte) {
	instance.lastActiveAt = time.Now()
}

func (n *NotifierHeartbeat) OnInstanceStderr(instance *PluginInstance, data []byte) {
	instance.lastActiveAt = time.Now()
}
