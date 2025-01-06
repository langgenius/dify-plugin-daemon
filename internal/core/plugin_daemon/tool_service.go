package plugin_daemon

import (
	"bytes"
	"encoding/base64"
	"errors"

	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/tool_entities"
	"github.com/xeipuuv/gojsonschema"
)

func InvokeTool(
	session *session_manager.Session,
	request *requests.RequestInvokeTool,
) (
	*stream.Stream[tool_entities.ToolResponseChunk], error,
) {
	runtime := session.Runtime()
	if runtime == nil {
		return nil, errors.New("plugin not found")
	}

	response, err := GenericInvokePlugin[
		requests.RequestInvokeTool, tool_entities.ToolResponseChunk,
	](
		session,
		request,
		128,
	)

	if err != nil {
		return nil, err
	}

	toolDeclaration := runtime.Configuration().Tool
	if toolDeclaration == nil {
		return nil, errors.New("tool declaration not found")
	}

	var toolOutputSchema plugin_entities.ToolOutputSchema
	for _, v := range toolDeclaration.Tools {
		if v.Identity.Name == request.Tool {
			toolOutputSchema = v.OutputSchema
		}
	}

	newResponse := stream.NewStream[tool_entities.ToolResponseChunk](128)
	routine.Submit(map[string]string{
		"module":        "plugin_daemon",
		"function":      "InvokeTool",
		"tool_name":     request.Tool,
		"tool_provider": request.Provider,
	}, func() {
		files := make(map[string]*bytes.Buffer)
		defer newResponse.Close()

		for response.Next() {
			item, err := response.Read()
			if err != nil {
				newResponse.WriteError(err)
				return
			}

			if item.Type == tool_entities.ToolResponseChunkTypeBlobChunk {
				id, ok := item.Message["id"].(string)
				if !ok {
					continue
				}

				totalLength, ok := item.Message["total_length"].(float64)
				if !ok {
					continue
				}

				// convert total_length to int
				totalLengthInt := int(totalLength)

				blob, ok := item.Message["blob"].(string)
				if !ok {
					continue
				}

				end, ok := item.Message["end"].(bool)
				if !ok {
					continue
				}

				if _, ok := files[id]; !ok {
					files[id] = bytes.NewBuffer(make([]byte, 0, totalLengthInt))
				}

				if end {
					newResponse.Write(tool_entities.ToolResponseChunk{
						Type: tool_entities.ToolResponseChunkTypeBlob,
						Message: map[string]any{
							"blob": files[id].Bytes(), // bytes will be encoded to base64 finally
						},
						Meta: item.Meta,
					})
				} else {
					if files[id].Len() > 15*1024*1024 {
						// delete the file if it is too large
						delete(files, id)
						newResponse.WriteError(errors.New("file is too large"))
						return
					} else {
						// decode the blob using base64
						decoded, err := base64.StdEncoding.DecodeString(blob)
						if err != nil {
							newResponse.WriteError(err)
							return
						}
						if len(decoded) > 8192 {
							// single chunk is too large, raises error
							newResponse.WriteError(errors.New("single file chunk is too large"))
							return
						}
						files[id].Write(decoded)
					}
				}
			} else {
				newResponse.Write(item)
			}
		}
	})

	// bind json schema validator
	bindToolValidator(response, toolOutputSchema)

	return newResponse, nil
}

func bindToolValidator(
	response *stream.Stream[tool_entities.ToolResponseChunk],
	toolOutputSchema plugin_entities.ToolOutputSchema,
) {
	// check if the tool_output_schema is valid
	variables := make(map[string]any)

	response.Filter(func(trc tool_entities.ToolResponseChunk) error {
		if trc.Type == tool_entities.ToolResponseChunkTypeVariable {
			variableName, ok := trc.Message["variable_name"].(string)
			if !ok {
				return errors.New("variable name is not a string")
			}
			stream, ok := trc.Message["stream"].(bool)
			if !ok {
				return errors.New("stream is not a boolean")
			}

			if stream {
				// ensure variable_value is a string
				variableValue, ok := trc.Message["variable_value"].(string)
				if !ok {
					return errors.New("variable value is not a string")
				}

				// create it if not exists
				if _, ok := variables[variableName]; !ok {
					variables[variableName] = ""
				}

				originalValue, ok := variables[variableName].(string)
				if !ok {
					return errors.New("variable value is not a string")
				}

				// add the variable value to the variable
				variables[variableName] = originalValue + variableValue
			} else {
				variables[variableName] = trc.Message["variable_value"]
			}
		}

		return nil
	})

	response.BeforeClose(func() {
		// validate the variables
		schema, err := gojsonschema.NewSchema(gojsonschema.NewGoLoader(toolOutputSchema))
		if err != nil {
			response.WriteError(err)
			return
		}

		// validate the variables
		result, err := schema.Validate(gojsonschema.NewGoLoader(variables))
		if err != nil {
			response.WriteError(err)
			return
		}

		if !result.Valid() {
			response.WriteError(errors.New("tool output schema is not valid"))
			return
		}
	})
}

func ValidateToolCredentials(
	session *session_manager.Session,
	request *requests.RequestValidateToolCredentials,
) (
	*stream.Stream[tool_entities.ValidateCredentialsResult], error,
) {
	return GenericInvokePlugin[requests.RequestValidateToolCredentials, tool_entities.ValidateCredentialsResult](
		session,
		request,
		1,
	)
}

func GetToolRuntimeParameters(
	session *session_manager.Session,
	request *requests.RequestGetToolRuntimeParameters,
) (
	*stream.Stream[tool_entities.GetToolRuntimeParametersResponse], error,
) {
	return GenericInvokePlugin[requests.RequestGetToolRuntimeParameters, tool_entities.GetToolRuntimeParametersResponse](
		session,
		request,
		1,
	)
}
