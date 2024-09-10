package service

import (
	"io"
	"mime/multipart"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager/installer"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

func InstallPlugin(c *gin.Context, dify_pkg_file multipart.File) {
	plugin_manager := plugin_manager.Manager()

	plugin_file, err := io.ReadAll(dify_pkg_file)
	if err != nil {
		c.JSON(200, entities.NewErrorResponse(-500, err.Error()))
		return
	}

	decoder, err := decoder.NewZipPluginDecoder(plugin_file)
	if err != nil {
		c.JSON(200, entities.NewErrorResponse(-500, err.Error()))
		return
	}

	baseSSEService(
		func() (*stream.Stream[installer.PluginInstallResponse], error) {
			return plugin_manager.Install(decoder)
		},
		c,
		3600,
	)
}
