package plugin_manager

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/serverless"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

type PluginInstallEvent string

const (
	PluginInstallEventInfo  PluginInstallEvent = "info"
	PluginInstallEventDone  PluginInstallEvent = "done"
	PluginInstallEventError PluginInstallEvent = "error"
)

type PluginInstallResponse struct {
	Event PluginInstallEvent `json:"event"`
	Data  string             `json:"data"`
}

// InstallToAWSFromPkg installs a plugin to AWS Lambda
func (p *PluginManager) InstallToAWSFromPkg(decoder decoder.PluginDecoder) (
	*stream.Stream[PluginInstallResponse], error,
) {
	response, err := serverless.UploadPlugin(decoder)
	if err != nil {
		return nil, err
	}

	new_response := stream.NewStream[PluginInstallResponse](2)
	routine.Submit(func() {
		response.Async(func(r serverless.LaunchAWSLambdaFunctionResponse) {
			if r.Event == serverless.Info {
				new_response.Write(PluginInstallResponse{
					Event: PluginInstallEventInfo,
					Data:  "Installing...",
				})
			} else if r.Event == serverless.Done {
				new_response.Write(PluginInstallResponse{
					Event: PluginInstallEventDone,
					Data:  "Installed",
				})
			} else if r.Event == serverless.Error {
				new_response.Write(PluginInstallResponse{
					Event: PluginInstallEventError,
				})
			}
		})
	})

	return new_response, nil
}

// InstallToLocal installs a plugin to local
func (p *PluginManager) InstallToLocal(decoder decoder.PluginDecoder) (
	*stream.Stream[PluginInstallResponse], error,
) {
	return nil, nil
}
