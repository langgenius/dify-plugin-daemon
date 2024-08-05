package server

import (
	"bytes"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
)

func CheckingKey(key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// get header X-Api-Key
		if c.GetHeader("X-Api-Key") != key {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		c.Next()
	}
}

type ginContextReader struct {
	reader *bytes.Reader
}

func (g *ginContextReader) Read(p []byte) (n int, err error) {
	return g.reader.Read(p)
}

func (g *ginContextReader) Close() error {
	return nil
}

// RedirectPluginInvoke redirects the request to the correct cluster node
func (app *App) RedirectPluginInvoke() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// get plugin identity
		raw, err := ctx.GetRawData()
		if err != nil {
			ctx.AbortWithStatusJSON(400, gin.H{"error": "Invalid request"})
			return
		}

		ctx.Request.Body = &ginContextReader{
			reader: bytes.NewReader(raw),
		}

		identity, err := parser.UnmarshalJsonBytes[plugin_entities.InvokePluginPluginIdentity](raw)

		if err != nil {
			ctx.AbortWithStatusJSON(400, gin.H{"error": "Invalid request"})
			return
		}

		plugin_id := parser.MarshalPluginIdentity(identity.PluginName, identity.PluginVersion)

		// check if plugin in current node
		if !app.cluster.IsPluginNoCurrentNode(
			plugin_id,
		) {
			// try find the correct node
			nodes, err := app.cluster.FetchPluginAvailableNodesById(plugin_id)
			if err != nil {
				ctx.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
				log.Error("fetch plugin available nodes failed: %s", err.Error())
				return
			} else if len(nodes) == 0 {
				ctx.AbortWithStatusJSON(404, gin.H{"error": "No available node"})
				log.Error("no available node")
				return
			}

			// redirect to the correct node
			node_id := nodes[0]
			status_code, header, body, err := app.cluster.RedirectRequest(node_id, ctx.Request)
			if err != nil {
				log.Error("redirect request failed: %s", err.Error())
				ctx.AbortWithStatusJSON(500, gin.H{"error": "Internal server error"})
				return
			}

			// set status code
			ctx.Writer.WriteHeader(status_code)

			// set header
			for key, values := range header {
				for _, value := range values {
					ctx.Writer.Header().Set(key, value)
				}
			}

			for {
				buf := make([]byte, 1024)
				n, err := body.Read(buf)
				if err != nil && err != io.EOF {
					break
				} else if err != nil {
					ctx.Writer.Write(buf[:n])
					break
				}

				if n > 0 {
					ctx.Writer.Write(buf[:n])
				}
			}

			ctx.Abort()
			return
		} else {
			ctx.Next()
		}
	}
}
