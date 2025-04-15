package plugin_daemon

import (
	"errors"

	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
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

	// bind json schema validator
	bindToolValidator(response, toolOutputSchema)

	return response, nil
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
