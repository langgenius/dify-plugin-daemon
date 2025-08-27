package serverless

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/http_requests"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

type ServerlessFunction struct {
	FunctionName string `json:"function_name" validate:"required"`
	FunctionDRN  string `json:"function_drn" validate:"required"`
	FunctionURL  string `json:"function_url" validate:"required"`
}

// Ping the serverless connector, return error if failed
func Ping() error {
	url, err := url.JoinPath(baseurl.String(), "/ping")
	if err != nil {
		return err
	}
	response, err := http_requests.PostAndParse[string](
		client,
		url,
		http_requests.HttpHeader(map[string]string{
			"Authorization": SERVERLESS_CONNECTOR_API_KEY,
		}),
	)
	if err != nil {
		return err
	}

	if response == nil || *response != "pong" {
		return fmt.Errorf("unexpected response from serverless connector: %s", *response)
	}

	return nil
}

var (
	ErrFunctionNotFound = errors.New("no function found")
)

// Fetch the function from serverless connector, return error if failed
func FetchFunction(manifest plugin_entities.PluginDeclaration, checksum string) (*ServerlessFunction, error) {
	filename := getFunctionFilename(manifest, checksum)

	url, err := url.JoinPath(baseurl.String(), "/v1/runner/instances")
	if err != nil {
		return nil, err
	}

	response, err := http_requests.GetAndParse[RunnerInstances](
		client,
		url,
		http_requests.HttpHeader(map[string]string{
			"Authorization": SERVERLESS_CONNECTOR_API_KEY,
		}),
		http_requests.HttpParams(map[string]string{
			"filename": filename,
		}),
	)

	if err != nil {
		return nil, err
	}

	if response.Error != "" {
		return nil, fmt.Errorf("unexpected response from plugin controller: %s", response.Error)
	}

	if len(response.Items) == 0 {
		return nil, ErrFunctionNotFound
	}

	return &ServerlessFunction{
		FunctionName: response.Items[0].Name,
		FunctionDRN:  response.Items[0].ResourceName,
		FunctionURL:  response.Items[0].Endpoint,
	}, nil
}

type LaunchFunctionEvent string

const (
	Error       LaunchFunctionEvent = "error"
	Info        LaunchFunctionEvent = "info"
	Function    LaunchFunctionEvent = "function"
	FunctionUrl LaunchFunctionEvent = "function_url"
	Done        LaunchFunctionEvent = "done"
)

type LaunchFunctionResponse struct {
	Event   LaunchFunctionEvent `json:"event"`
	Message string              `json:"message"`
}

// Setup the function from serverless connector, it will receive the context as the input
// and build it a docker image, then run it on serverless platform like AWS Lambda
// it returns a event stream, the caller should consider it as a async operation
func SetupFunction(
	pluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier,
	manifest plugin_entities.PluginDeclaration,
	checksum string,
	context io.Reader,
	timeout int, // in seconds
) (*stream.Stream[LaunchFunctionResponse], error) {
	url, err := url.JoinPath(baseurl.String(), "/v1/launch")
	if err != nil {
		return nil, err
	}

	metadata, err := json.Marshal(map[string]string{
		"plugin_unique_identifier": pluginUniqueIdentifier.String(),
	})
	if err != nil {
		return nil, err
	}

	// join a filename
	serverless_connector_response, err := http_requests.PostAndParseStream[LaunchFunctionResponseChunk](
		client,
		url,
		http_requests.HttpHeader(map[string]string{
			"Authorization": SERVERLESS_CONNECTOR_API_KEY,
		}),
		http_requests.HttpReadTimeout(int64(timeout)*1000),
		http_requests.HttpWriteTimeout(int64(timeout)*1000),
		http_requests.HttpPayloadMultipart(
			map[string]string{
				"verified": func() string {
					if manifest.Verified {
						return "true"
					}
					return "false"
				}(),
				"metadata": string(metadata),
			},
			map[string]http_requests.HttpPayloadMultipartFile{
				"context": {
					Filename: getFunctionFilename(manifest, checksum),
					Reader:   context,
				},
			},
		),
	)
	if err != nil {
		return nil, err
	}

	response := stream.NewStream[LaunchFunctionResponse](10)

	routine.Submit(map[string]string{
		"module": "serverless_connector",
		"func":   "SetupFunction",
	}, func() {
		defer response.Close()
		if err := serverless_connector_response.Async(func(chunk LaunchFunctionResponseChunk) {
			if chunk.State == LAUNCH_STATE_FAILED {
				response.Write(LaunchFunctionResponse{
					Event:   Error,
					Message: chunk.Message,
				})
				return
			}

			switch chunk.Stage {
			case LAUNCH_STAGE_START, LAUNCH_STAGE_BUILD:
				response.Write(LaunchFunctionResponse{
					Event:   Info,
					Message: "Building plugin...",
				})
			case LAUNCH_STAGE_RUN:
				if chunk.State == LAUNCH_STATE_SUCCESS {
					data, err := parser.ParserCommaSeparatedValues[LaunchFunctionFinalStageMessage]([]byte(chunk.Message))
					if err != nil {
						response.Write(LaunchFunctionResponse{
							Event:   Error,
							Message: err.Error(),
						})
						return
					}

					response.Write(LaunchFunctionResponse{
						Event:   Function,
						Message: data.Name,
					})
					response.Write(LaunchFunctionResponse{
						Event:   FunctionUrl,
						Message: data.Endpoint,
					})
				} else {
					response.Write(LaunchFunctionResponse{
						Event:   Info,
						Message: "Launching plugin...",
					})
				}
			case LAUNCH_STAGE_END:
				response.Write(LaunchFunctionResponse{
					Event:   Done,
					Message: "Plugin launched",
				})
			}
		}); err != nil {
			response.Write(LaunchFunctionResponse{
				Event:   Error,
				Message: err.Error(),
			})
		}
	})

	return response, nil
}
