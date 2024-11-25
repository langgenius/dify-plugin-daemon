package plugin_manager

import (
	"fmt"
	"regexp"

	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/serverless"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/db"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

var (
	variablePattern = regexp.MustCompile(`(\w+)=([^,]+)`)
)

func extractVariables(message string) map[string]string {
	matches := variablePattern.FindAllStringSubmatch(message, -1)
	variables := make(map[string]string)
	for _, match := range matches {
		variables[match[1]] = match[2]
	}
	return variables
}

// InstallToAWSFromPkg installs a plugin to AWS Lambda
func (p *PluginManager) InstallToAWSFromPkg(
	decoder decoder.PluginDecoder,
	source string,
	meta map[string]any,
) (
	*stream.Stream[PluginInstallResponse], error,
) {
	checksum, err := decoder.Checksum()
	if err != nil {
		return nil, err
	}
	declaration, err := decoder.Manifest()
	if err != nil {
		return nil, err
	}
	uniqueIdentity, err := decoder.UniqueIdentity()
	if err != nil {
		return nil, err
	}

	response, err := serverless.UploadPlugin(decoder)
	if err != nil {
		return nil, err
	}

	newResponse := stream.NewStream[PluginInstallResponse](128)
	routine.Submit(func() {
		defer func() {
			newResponse.Close()
		}()

		lambdaUrl := ""
		lambdaFunctionName := ""

		response.Async(func(r serverless.LaunchAWSLambdaFunctionResponse) {
			if r.Stage == serverless.LaunchStageStart {
				newResponse.Write(PluginInstallResponse{
					Event: PluginInstallEventInfo,
					Data:  "Installing...",
				})
			} else if r.Stage == serverless.LaunchStageRun {
				if r.State == serverless.LaunchStateSuccess {
					// split the message to get the lambda url and function name
					variables := extractVariables(r.Message)
					if len(variables) != 3 {
						newResponse.Write(PluginInstallResponse{
							Event: PluginInstallEventError,
							Data:  "Internal server error, failed to get lambda url or function name",
						})
						return
					}

					lambdaUrl = variables["endpoint"]
					lambdaFunctionName = variables["name"]

					if lambdaUrl == "" || lambdaFunctionName == "" {
						newResponse.Write(PluginInstallResponse{
							Event: PluginInstallEventError,
							Data: fmt.Sprintf(
								"Internal server error, failed to get lambda url or function name,"+
									" one of them is empty, lambdaUrl: %3s, lambdaFunctionName: %3s",
								lambdaUrl, lambdaFunctionName,
							),
						})
						return
					}

					// check if the plugin is already installed
					_, err := db.GetOne[models.ServerlessRuntime](
						db.Equal("checksum", checksum),
						db.Equal("type", string(models.SERVERLESS_RUNTIME_TYPE_AWS_LAMBDA)),
					)
					if err == db.ErrDatabaseNotFound {
						// create a new serverless runtime
						serverlessModel := &models.ServerlessRuntime{
							Checksum:               checksum,
							Type:                   models.SERVERLESS_RUNTIME_TYPE_AWS_LAMBDA,
							FunctionURL:            lambdaUrl,
							FunctionName:           lambdaFunctionName,
							PluginUniqueIdentifier: uniqueIdentity.String(),
							Declaration:            declaration,
						}
						err = db.Create(serverlessModel)
						if err != nil {
							newResponse.Write(PluginInstallResponse{
								Event: PluginInstallEventError,
								Data:  "Failed to create serverless runtime",
							})
							return
						}
					} else if err != nil {
						newResponse.Write(PluginInstallResponse{
							Event: PluginInstallEventError,
							Data:  "Failed to check if the plugin is already installed",
						})
						return
					}

				} else if r.State == serverless.LaunchStateRunning {
					// do nothing
				}
			} else if r.Stage == serverless.LaunchStageEnd {
				if r.State == serverless.LaunchStateSuccess {
					newResponse.Write(PluginInstallResponse{
						Event: PluginInstallEventDone,
						Data:  "Installed",
					})
				}
			} else if r.Stage == serverless.LaunchStageBuild {
				// do nothing
				newResponse.Write(PluginInstallResponse{
					Event: PluginInstallEventInfo,
					Data:  "Building...",
				})
			}

			// any error occurs
			if r.State == serverless.LaunchStateFailed {
				newResponse.Write(PluginInstallResponse{
					Event: PluginInstallEventError,
					Data:  "Internal server error",
				})
				// log the error message
				log.Error("failed to install plugin to AWS Lambda, stage: %s, state: %s, error: %s", r.Stage, r.State, r.Message)
			}
		})
	})

	return newResponse, nil
}
