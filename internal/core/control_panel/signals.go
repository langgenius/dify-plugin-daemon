package controlpanel

type ControlPanelSignal string

const (
	// A new plugin instance is starting
	CONTROL_SIGNAL_NEW_INSTANCE_STARTING ControlPanelSignal = "new_instance_starting"

	// A new plugin instance is ready to receive requests
	CONTROL_SIGNAL_NEW_INSTANCE_READY ControlPanelSignal = "new_instance_ready"

	// Control panel failed to launch a new plugin instance
	CONTROL_SIGNAL_NEW_INSTANCE_FAILED ControlPanelSignal = "new_instance_failed"

	// A plugin instance is shutting down
	CONTROL_SIGNAL_INSTANCE_SHUTDOWN ControlPanelSignal = "instance_shutdown"
)

type InstanceSignal string

const (
	// A plugin instance is starting
	INSTANCE_SIGNAL_STARTING InstanceSignal = "instance_starting"

	// A plugin instance is ready to receive requests
	INSTANCE_SIGNAL_READY InstanceSignal = "instance_ready"

	// A plugin instance failed to start
	INSTANCE_SIGNAL_FAILED InstanceSignal = "instance_failed"

	// A plugin instance is shutting down
	INSTANCE_SIGNAL_SHUTDOWN InstanceSignal = "instance_shutdown"
)
