package service

import (
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel"
	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/model_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache/helper"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/stream"
)

func GetAIModelSchema(
	r *plugin_entities.InvokePluginRequest[requests.RequestGetAIModelSchema],
	ctx *gin.Context,
	max_timeout_seconds int,
) {
	identity := helper.ModelSchemaCacheIdentity{
		UniqueIdentifier: r.UniqueIdentifier,
		ModelType:        string(r.Data.ModelType),
		Model:            r.Data.Model,
		CredentialType:   r.Data.Credentials.CredentialType,
		Credentials:      r.Data.Credentials.Credentials,
	}

	if cached := helper.GetCachedModelSchema(identity); cached != nil {
		writeSingleSSEChunk(ctx, *cached)
		return
	}

	baseSSEWithSession(
		func(session *session_manager.Session) (*stream.Stream[model_entities.GetModelSchemasResponse], error) {
			response, err := io_tunnel.GenericInvokePlugin[
				requests.RequestGetAIModelSchema,
				model_entities.GetModelSchemasResponse,
			](session, &r.Data, 1)
			if err != nil {
				return nil, err
			}

			cacheModelSchemaStream(response, identity)
			return response, nil
		},
		access_types.PLUGIN_ACCESS_TYPE_MODEL,
		access_types.PLUGIN_ACCESS_ACTION_GET_AI_MODEL_SCHEMAS,
		r,
		ctx,
		max_timeout_seconds,
	)
}

// The plugin answers with a single chunk. Anything else, including an error mid-stream,
// leaves the cache untouched rather than storing a partial answer.
func cacheModelSchemaStream(
	response *stream.Stream[model_entities.GetModelSchemasResponse],
	identity helper.ModelSchemaCacheIdentity,
) {
	var chunks []model_entities.GetModelSchemasResponse

	response.Filter(func(chunk model_entities.GetModelSchemasResponse) error {
		chunks = append(chunks, chunk)
		return nil
	})

	response.OnClose(func() {
		if len(chunks) != 1 || chunks[0].ModelSchema == nil {
			return
		}
		helper.StoreModelSchema(identity, chunks[0])
	})
}

// Mirrors the wire format baseSSEService writes, so a cache hit is indistinguishable to the
// caller from a plugin invocation.
func writeSingleSSEChunk(ctx *gin.Context, chunk model_entities.GetModelSchemasResponse) {
	writer := ctx.Writer
	writer.WriteHeader(200)
	writer.Header().Set("Content-Type", "text/event-stream")

	writer.Write([]byte("data: "))
	writer.Write(parser.MarshalJsonBytes(entities.NewSuccessResponse(chunk)))
	writer.Write([]byte("\n\n"))
	writer.Flush()
}
