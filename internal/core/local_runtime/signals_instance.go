package local_runtime

type InstanceSignal string

const (
	// A plugin instance is starting
	INSTANCE_SIGNAL_STARTING InstanceSignal = "instance_starting"

	// A plugin instance is ready to receive requests
	INSTANCE_SIGNAL_READY InstanceSignal = "instance_ready"

	// A plugin instance failed to start
	INSTANCE_SIGNAL_FAILED InstanceSignal = "instance_failed"

	// A plugin sent a log message
	INSTANCE_SIGNAL_LOG InstanceSignal = "instance_log"

	// A plugin sent an error message
	INSTANCE_SIGNAL_ERROR InstanceSignal = "instance_error"

	// A plugin instance is shutting down
	INSTANCE_SIGNAL_SHUTDOWN InstanceSignal = "instance_shutdown"

	// A plugin instance is sending a heartbeat
	INSTANCE_SIGNAL_HEARTBEAT InstanceSignal = "instance_heartbeat"
)

type InstanceSignalEntity struct {
	// The signal type
	Signal InstanceSignal
}

type PluginInstanceNotifier interface {
	// on instance starting
	OnInstanceStarting()

	// on instance ready
	OnInstanceReady(*PluginInstance)

	// on instance failed
	OnInstanceLaunchFailed(*PluginInstance, error)

	// on instance shutdown
	OnInstanceShutdown(*PluginInstance)

	// on instance heartbeat
	OnInstanceHeartbeat(*PluginInstance)

	// on instance log
	OnInstanceLog(*PluginInstance, string)

	// on instance error
	OnInstanceErrorLog(*PluginInstance, error)

	// on instance warning message
	OnInstanceWarningLog(*PluginInstance, string)

	// on instance stdout
	OnInstanceStdout(*PluginInstance, []byte)

	// on instance stderr
	OnInstanceStderr(*PluginInstance, []byte)
}

// AddNotifier adds a notifier layer to the plugin instance
func (s *PluginInstance) AddNotifier(notifier PluginInstanceNotifier) {
	s.notifierLock.Lock()
	defer s.notifierLock.Unlock()
	s.notifiers = append(s.notifiers, notifier)
}

// WalkNotifiers walks through all notifiers and calls the corresponding method
func (s *PluginInstance) WalkNotifiers(
	callback func(notifier PluginInstanceNotifier),
) {
	s.notifierLock.Lock()
	notifiers := s.notifiers // copy notifiers, prevent race condition access
	s.notifierLock.Unlock()

	for _, notifier := range notifiers {
		callback(notifier)
	}
}
