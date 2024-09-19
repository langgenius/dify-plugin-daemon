package remote_manager

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

func (r *RemotePluginRuntime) Identity() (plugin_entities.PluginUniqueIdentifier, error) {
	// copy a new declaration
	config := r.Config
	config.Author = r.tenant_id
	checksum, _ := r.Checksum()
	return plugin_entities.NewPluginUniqueIdentifier(fmt.Sprintf("%s@%s", config.Identity(), checksum))
}

func (r *RemotePluginRuntime) Cleanup() {
	// no cleanup needed
}

func (r *RemotePluginRuntime) WaitStarted() <-chan bool {
	r.wait_chan_lock.Lock()
	defer r.wait_chan_lock.Unlock()

	ch := make(chan bool)
	r.wait_started_chan = append(r.wait_started_chan, ch)
	return ch
}

func (r *RemotePluginRuntime) WaitStopped() <-chan bool {
	r.wait_chan_lock.Lock()
	defer r.wait_chan_lock.Unlock()

	ch := make(chan bool)
	r.wait_stopped_chan = append(r.wait_stopped_chan, ch)
	return ch
}
