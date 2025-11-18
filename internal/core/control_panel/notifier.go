package controlpanel

/*
ControlPanelNotifier is a interface that can be used to notify the control panel
about the plugin runtime.
*/

func (c *ControlPanel) AddNotifier(notifier ControlPanelNotifier) {
	c.controlPanelNotifierLock.Lock()
	defer c.controlPanelNotifierLock.Unlock()
	c.controlPanelNotifiers = append(c.controlPanelNotifiers, notifier)
}

func (c *ControlPanel) WalkNotifiers(fn func(notifier ControlPanelNotifier)) {
	c.controlPanelNotifierLock.RLock()
	notifiers := c.controlPanelNotifiers // copy the notifiers
	c.controlPanelNotifierLock.RUnlock()

	for _, notifier := range notifiers {
		fn(notifier)
	}
}
