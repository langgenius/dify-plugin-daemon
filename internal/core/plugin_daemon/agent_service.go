package plugin_daemon

import (
	"bytes"
	"encoding/base64"
	"errors"

	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/agent_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/tool_entities"
)

func InvokeAgentStrategy(
	session *session_manager.Session,
	r *requests.RequestInvokeAgentStrategy,
) (*stream.Stream[agent_entities.AgentStrategyResponseChunk], error) {
	runtime := session.Runtime()
	if runtime == nil {
		return nil, errors.New("plugin not found")
	}

	response, err := GenericInvokePlugin[
		requests.RequestInvokeAgentStrategy, agent_entities.AgentStrategyResponseChunk,
	](
		session,
		r,
		128,
	)

	if err != nil {
		return nil, err
	}

	newResponse := stream.NewStream[agent_entities.AgentStrategyResponseChunk](128)
	files := make(map[string]*bytes.Buffer)

	routine.Submit(map[string]string{
		"agent_service": "invoke_agent_strategy",
	}, func() {
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
					newResponse.Write(agent_entities.AgentStrategyResponseChunk{
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

	return newResponse, nil
}
