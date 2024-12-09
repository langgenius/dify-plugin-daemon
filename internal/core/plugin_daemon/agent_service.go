package plugin_daemon

import (
	"bytes"
	"encoding/base64"
	"errors"

	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/agent_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/tool_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/xeipuuv/gojsonschema"
)

func InvokeAgent(
	session *session_manager.Session,
	r *requests.RequestInvokeAgent,
) (*stream.Stream[agent_entities.AgentResponseChunk], error) {
	runtime := session.Runtime()
	if runtime == nil {
		return nil, errors.New("plugin not found")
	}

	response, err := GenericInvokePlugin[
		requests.RequestInvokeAgent, agent_entities.AgentResponseChunk,
	](
		session,
		r,
		128,
	)

	if err != nil {
		return nil, err
	}

	agentDeclaration := runtime.Configuration().Agent
	if agentDeclaration == nil {
		return nil, errors.New("agent declaration not found")
	}

	var agentOutputSchema plugin_entities.AgentOutputSchema
	for _, v := range agentDeclaration.Strategies {
		if v.Identity.Name == r.Strategy {
			agentOutputSchema = v.OutputSchema
		}
	}

	newResponse := stream.NewStream[agent_entities.AgentResponseChunk](128)
	routine.Submit(map[string]string{
		"module":         "plugin_daemon",
		"function":       "InvokeAgent",
		"agent_name":     r.Strategy,
		"agent_provider": r.Provider,
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
					newResponse.Write(agent_entities.AgentResponseChunk{
						ToolResponseChunk: tool_entities.ToolResponseChunk{
							Type: tool_entities.ToolResponseChunkTypeBlob,
							Message: map[string]any{
								"blob": files[id].Bytes(), // bytes will be encoded to base64 finally
							},
							Meta: item.Meta,
						},
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
	bindAgentValidator(response, agentOutputSchema)

	return newResponse, nil
}

// TODO: reduce implementation of bindAgentValidator, it's a copy of bindToolValidator now
func bindAgentValidator(
	response *stream.Stream[agent_entities.AgentResponseChunk],
	agentOutputSchema plugin_entities.AgentOutputSchema,
) {
	// check if the tool_output_schema is valid
	variables := make(map[string]any)

	response.Filter(func(trc agent_entities.AgentResponseChunk) error {
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
		schema, err := gojsonschema.NewSchema(gojsonschema.NewGoLoader(agentOutputSchema))
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
