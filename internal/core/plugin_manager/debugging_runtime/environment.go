package debugging_runtime

import (
	"fmt"
	"regexp"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

var (
	authorRegex     = regexp.MustCompile(`^[a-z0-9_-]{1,64}$`)
	pluginNameRegex = regexp.MustCompile(`^[a-z0-9_-]{1,64}$`)
)

func (r *RemotePluginRuntime) Identity() (plugin_entities.PluginUniqueIdentifier, error) {
	// copy a new declaration
	// check original author is alphanumeric
	if !authorRegex.MatchString(r.Config.Author) {
		return "", fmt.Errorf("author must be alphanumeric and less than 64 characters: ^[a-z0-9_-]{1,64}$")
	}
	if !pluginNameRegex.MatchString(r.Config.Name) {
		return "", fmt.Errorf("plugin name must be alphanumeric and less than 64 characters: ^[a-z0-9_-]{1,64}$")
	}
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
