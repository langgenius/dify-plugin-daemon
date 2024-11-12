package remote_manager

import (
	"fmt"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

func (r *RemotePluginRuntime) Identity() (plugin_entities.PluginUniqueIdentifier, error) {
	// copy a new declaration
	config := r.Config
	config.Author = r.tenantId
	checksum, _ := r.Checksum()
	return plugin_entities.NewPluginUniqueIdentifier(fmt.Sprintf("%s@%s", config.Identity(), checksum))
}

func (r *RemotePluginRuntime) Cleanup() {
	// no cleanup needed
}

func (r *RemotePluginRuntime) WaitStarted() <-chan bool {
	r.waitChanLock.Lock()
	defer r.waitChanLock.Unlock()

	ch := make(chan bool)
	r.waitStartedChan = append(r.waitStartedChan, ch)
	return ch
}

func (r *RemotePluginRuntime) WaitStopped() <-chan bool {
	r.waitChanLock.Lock()
	defer r.waitChanLock.Unlock()

	ch := make(chan bool)
	r.waitStoppedChan = append(r.waitStoppedChan, ch)
	return ch
}
