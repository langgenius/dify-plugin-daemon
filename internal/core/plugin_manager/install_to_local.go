package plugin_manager

import (
	"time"

	"errors"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

// InstallToLocal installs a plugin to local
func (p *PluginManager) InstallToLocal(
    plugin_unique_identifier plugin_entities.PluginUniqueIdentifier,
    source string,
    meta map[string]any,
    options InstallOptions,
) (
    *stream.Stream[PluginInstallResponse], error,
) {
    log.Info("[install] start InstallToLocal: uid=%s source=%s blue_green=%v", plugin_unique_identifier.String(), source, options.BlueGreen)
    packageFile, err := p.packageBucket.Get(plugin_unique_identifier.String())
    if err != nil {
        return nil, err
    }

	err = p.installedBucket.Save(plugin_unique_identifier, packageFile)
	if err != nil {
		return nil, err
	}

	runtime, _, errChan, err := p.launchLocal(plugin_unique_identifier, options)
	if err != nil {
		return nil, err
	}

	response := stream.NewStream[PluginInstallResponse](128)
	routine.Submit(map[string]string{
		"module":   "plugin_manager",
		"function": "InstallToLocal",
	}, func() {
		defer response.Close()

		ticker := time.NewTicker(time.Second * 5) // check heartbeat every 5 seconds
		defer ticker.Stop()
		timer := time.NewTimer(time.Second * 240) // timeout after 240 seconds
		defer timer.Stop()

		// wait for plugin to fully start to ensure traffic cutover is safe
        startedCh := runtime.WaitStarted()

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
            case err := <-errChan:
                if err != nil {
                    log.Error("[install] failed during launching: uid=%s err=%s", plugin_unique_identifier.String(), err.Error())
                    // if error occurs, delete the plugin from local and stop the plugin
                    identity, er := runtime.Identity()
					if er != nil {
						log.Error("get plugin identity failed: %s", er.Error())
					}
					if er := p.installedBucket.Delete(identity); er != nil {
						log.Error("delete plugin from local failed: %s", er.Error())
					}

					var errorMsg string
					if er != nil {
						errorMsg = errors.Join(err, er).Error()
					} else {
						errorMsg = err.Error()
					}

					response.Write(PluginInstallResponse{
						Event: PluginInstallEventError,
						Data:  errorMsg,
					})
					runtime.Stop()
					return
				}
            case <-startedCh:
                log.Info("[install] runtime started: uid=%s", plugin_unique_identifier.String())
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
