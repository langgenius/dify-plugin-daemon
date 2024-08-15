package remote_manager

import (
	"strings"
)

func (r *RemotePluginRuntime) Identity() (string, error) {
	return strings.Join([]string{r.Configuration().Identity(), r.tenant_id}, ":"), nil
}

func (r *RemotePluginRuntime) Cleanup() {
	// no cleanup needed
}
