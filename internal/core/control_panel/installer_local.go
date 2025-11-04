package controlpanel

import (
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

// InstallToLocalFromPkg installs a plugin to the local plugin runtime
// It's scope only for initializing environment, !!!starting schedule loop is not included
func (c *ControlPanel) InstallToLocalFromPkg(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
) (*stream.Stream[InstallLocalPluginResponse], error) {
	runtime, err := c.getLocalPluginRuntime(pluginUniqueIdentifier)
	if err != nil {
		return nil, err
	}

	response := stream.NewStream[InstallLocalPluginResponse](128)

	routine.Submit(map[string]string{
		"module": "installer_local",
		"func":   "InstallToLocalFromPkg",
	}, func() {
		defer response.Close()

		// write the initial info
		response.Write(InstallLocalPluginResponse{
			Event:   Info,
			Message: "Installing plugin to local runtime",
		})

		// init environment, create a channel to handle heartbeat to avoid network timeout
		c := make(chan error)
		ticker := time.NewTicker(5 * time.Second)

		routine.Submit(map[string]string{
			"module": "installer_local",
			"func":   "InstallToLocalFromPkg",
		}, func() {
			defer close(c)
			if err := runtime.InitEnvironment(); err != nil {
				c <- err
			}
		})

		// wait for the plugin to be installed
		for {
			select {
			case <-ticker.C:
				response.Write(InstallLocalPluginResponse{
					Event:   Info,
					Message: "Installing plugin to local runtime in progress...",
				})
			case err := <-c:
				if err == nil {
					response.Write(InstallLocalPluginResponse{
						Event:   Done,
						Message: "Plugin installed successfully",
					})
				} else {
					response.Write(InstallLocalPluginResponse{
						Event:   Error,
						Message: err.Error(),
					})
				}
				return
			}
		}
	})

	return response, nil
}
