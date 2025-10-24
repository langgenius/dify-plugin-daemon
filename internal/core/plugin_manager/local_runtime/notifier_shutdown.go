package local_runtime

type NotifierShutdown struct {
	callbacks []func()
}

func (s *NotifierShutdown) OnInstanceStarting(instance *PluginInstance) {

}

func (s *NotifierShutdown) OnInstanceReady(instance *PluginInstance) {

}

func (s *NotifierShutdown) OnInstanceFailed(instance *PluginInstance, err error) {

}

func (s *NotifierShutdown) OnInstanceShutdown(instance *PluginInstance) {
	for _, callback := range s.callbacks {
		callback()
	}
}

func (s *NotifierShutdown) OnInstanceHeartbeat(instance *PluginInstance) {

}

func (s *NotifierShutdown) OnInstanceLog(instance *PluginInstance, message string) {

}

func (s *NotifierShutdown) OnInstanceErrorLog(instance *PluginInstance, err error) {

}

func (s *NotifierShutdown) OnInstanceWarningLog(instance *PluginInstance, message string) {

}

func (s *NotifierShutdown) OnInstanceStdout(instance *PluginInstance, data []byte) {

}

func (s *NotifierShutdown) OnInstanceStderr(instance *PluginInstance, data []byte) {

}
