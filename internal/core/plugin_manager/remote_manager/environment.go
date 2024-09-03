package remote_manager

import (
	"fmt"
	"strings"
)

func (r *RemotePluginRuntime) Identity() (string, error) {
	identity := strings.Join([]string{r.Configuration().Identity(), r.tenant_id}, ":")
	return fmt.Sprintf("%s@%s", identity, r.Checksum()), nil
}

func (r *RemotePluginRuntime) Cleanup() {
	// no cleanup needed
}
