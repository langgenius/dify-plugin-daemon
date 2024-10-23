package plugin_daemon

import (
	"bytes"
	"encoding/base64"
	"errors"

	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/tool_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
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

	tool_declaration := runtime.Configuration().Tool
	if tool_declaration == nil {
		return nil, errors.New("tool declaration not found")
	}

	var tool_output_schema plugin_entities.ToolOutputSchema
	for _, v := range tool_declaration.Tools {
		if v.Identity.Name == request.Tool {
			tool_output_schema = v.OutputSchema
		}
	}

	new_response := stream.NewStream[tool_entities.ToolResponseChunk](128)
	routine.Submit(func() {
		files := make(map[string]*bytes.Buffer)
		defer new_response.Close()

		for response.Next() {
			item, err := response.Read()
			if err != nil {
				return
			}

			if item.Type == tool_entities.ToolResponseChunkTypeBlobChunk {
				id, ok := item.Message["id"].(string)
				if !ok {
					continue
				}

				total_length, ok := item.Message["total_length"].(float64)
				if !ok {
					continue
				}

				// convert total_length to int
				total_length_int := int(total_length)

				blob, ok := item.Message["blob"].(string)
				if !ok {
					continue
				}

				end, ok := item.Message["end"].(bool)
				if !ok {
					continue
				}

				if _, ok := files[id]; !ok {
					files[id] = bytes.NewBuffer(make([]byte, 0, total_length_int))
				}

				if end {
					new_response.Write(tool_entities.ToolResponseChunk{
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
						new_response.WriteError(errors.New("file is too large"))
						return
					} else {
						// decode the blob using base64
						decoded, err := base64.StdEncoding.DecodeString(blob)
						if err != nil {
							new_response.WriteError(err)
							return
						}
						if len(decoded) > 8192 {
							// single chunk is too large, raises error
							new_response.WriteError(errors.New("single file chunk is too large"))
							return
						}
						files[id].Write(decoded)
					}
				}
			} else {
				new_response.Write(item)
			}
		}
	})

	// bind json schema validator
	bindValidator(response, tool_output_schema)

	return new_response, nil
}

func bindValidator(
	response *stream.Stream[tool_entities.ToolResponseChunk],
	tool_output_schema plugin_entities.ToolOutputSchema,
) {
	// check if the tool_output_schema is valid
	variables := make(map[string]any)

	response.Filter(func(trc tool_entities.ToolResponseChunk) error {
		if trc.Type == tool_entities.ToolResponseChunkTypeVariable {
			variable_name, ok := trc.Message["variable_name"].(string)
			if !ok {
				return errors.New("variable name is not a string")
			}
			stream, ok := trc.Message["stream"].(bool)
			if !ok {
				return errors.New("stream is not a boolean")
			}

			if stream {
				// ensure variable_value is a string
				variable_value, ok := trc.Message["variable_value"].(string)
				if !ok {
					return errors.New("variable value is not a string")
				}

				// create it if not exists
				if _, ok := variables[variable_name]; !ok {
					variables[variable_name] = ""
				}

				original_value, ok := variables[variable_name].(string)
				if !ok {
					return errors.New("variable value is not a string")
				}

				// add the variable value to the variable
				variables[variable_name] = original_value + variable_value
			} else {
				variables[variable_name] = trc.Message["variable_value"]
			}
		}

		return nil
	})

	response.BeforeClose(func() {
		// validate the variables
		schema, err := gojsonschema.NewSchema(gojsonschema.NewGoLoader(tool_output_schema))
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
