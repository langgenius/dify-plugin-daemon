package plugin_manager

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

// InstallToLocal installs a plugin to local
func (p *PluginManager) InstallToLocal(
	plugin_path string,
	plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
	source string,
	meta map[string]any,
) (
	*stream.Stream[PluginInstallResponse], error,
) {
	plugin_file, err := os.Open(plugin_path)
	if err != nil {
		return nil, err
	}
	defer plugin_file.Close()
	installed_file_path := filepath.Join(p.pluginStoragePath, plugin_unique_identifier.String())
	dir_path := filepath.Dir(installed_file_path)
	if err := os.MkdirAll(dir_path, 0755); err != nil {
		return nil, err
	}
	installed_file, err := os.Create(installed_file_path)
	if err != nil {
		return nil, err
	}
	defer installed_file.Close()

	if _, err := io.Copy(installed_file, plugin_file); err != nil {
		return nil, err
	}

	runtime, launched_chan, err := p.launchLocal(installed_file_path)
	if err != nil {
		return nil, err
	}

	response := stream.NewStream[PluginInstallResponse](128)
	routine.Submit(func() {
		defer response.Close()

		ticker := time.NewTicker(time.Second * 5) // check heartbeat every 5 seconds
		defer ticker.Stop()
		timer := time.NewTimer(time.Second * 240) // timeout after 240 seconds
		defer timer.Stop()

		for {
			select {
			case <-ticker.C:
				// heartbeat
				response.Write(PluginInstallResponse{
					Event: PluginInstallEventInfo,
					Data:  "Installing",
				})
			case <-timer.C:
				// timeout
				response.Write(PluginInstallResponse{
					Event: PluginInstallEventInfo,
					Data:  "Timeout",
				})
				runtime.Stop()
				return
			case <-launched_chan:
				// launched
				if err != nil {
					response.Write(PluginInstallResponse{
						Event: PluginInstallEventError,
						Data:  err.Error(),
					})
					runtime.Stop()
					return
				}
				response.Write(PluginInstallResponse{
					Event: PluginInstallEventDone,
					Data:  "Installed",
				})
				return
			}
		}

	})

	return response, nil
}
