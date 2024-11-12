package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/service"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
	"github.com/langgenius/dify-plugin-daemon/internal/types/entities/plugin_entities"
)

func GetAsset(c *gin.Context) {
	plugin_manager := plugin_manager.Manager()
	asset, err := plugin_manager.GetAsset(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", asset)
}

func UploadPlugin(app *app.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		dify_pkg_file_header, err := c.FormFile("dify_pkg")
		if err != nil {
			c.JSON(http.StatusOK, entities.NewErrorResponse(-400, err.Error()))
			return
		}

		tenant_id := c.Param("tenant_id")
		if tenant_id == "" {
			c.JSON(http.StatusOK, entities.NewErrorResponse(-400, "Tenant ID is required"))
			return
		}

		if dify_pkg_file_header.Size > app.MaxPluginPackageSize {
			c.JSON(http.StatusOK, entities.NewErrorResponse(-413, "File size exceeds the maximum limit"))
			return
		}

		verify_signature := c.PostForm("verify_signature") == "true"

		dify_pkg_file, err := dify_pkg_file_header.Open()
		if err != nil {
			c.JSON(http.StatusOK, entities.NewErrorResponse(-400, err.Error()))
			return
		}
		defer dify_pkg_file.Close()

		c.JSON(http.StatusOK, service.UploadPluginFromPkg(app, c, tenant_id, dify_pkg_file, verify_signature))
	}
}

func UpgradePlugin(app *app.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		BindRequest(c, func(request struct {
			TenantID                       string                                 `uri:"tenant_id" validate:"required"`
			OriginalPluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier `json:"original_plugin_unique_identifier" validate:"required,plugin_unique_identifier"`
			NewPluginUniqueIdentifier      plugin_entities.PluginUniqueIdentifier `json:"new_plugin_unique_identifier" validate:"required,plugin_unique_identifier"`
			Source                         string                                 `json:"source" validate:"required"`
			Meta                           map[string]any                         `json:"meta" validate:"omitempty"`
		}) {
			c.JSON(http.StatusOK, service.UpgradePlugin(
				app,
				request.TenantID,
				request.Source,
				request.Meta,
				request.OriginalPluginUniqueIdentifier,
				request.NewPluginUniqueIdentifier,
			))
		})
	}
}

func InstallPluginFromIdentifiers(app *app.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		BindRequest(c, func(request struct {
			TenantID                string                                   `uri:"tenant_id" validate:"required"`
			PluginUniqueIdentifiers []plugin_entities.PluginUniqueIdentifier `json:"plugin_unique_identifiers" validate:"required,max=64,dive,plugin_unique_identifier"`
			Source                  string                                   `json:"source" validate:"required"`
			Meta                    map[string]any                           `json:"meta" validate:"omitempty"`
		}) {
			if request.Meta == nil {
				request.Meta = map[string]any{}
			}
			c.JSON(http.StatusOK, service.InstallPluginFromIdentifiers(
				app, request.TenantID, request.PluginUniqueIdentifiers, request.Source, request.Meta,
			))
		})
	}
}

func FetchPluginInstallationTasks(c *gin.Context) {
	BindRequest(c, func(request struct {
		TenantID string `uri:"tenant_id" validate:"required"`
		Page     int    `form:"page" validate:"required,min=1"`
		PageSize int    `form:"page_size" validate:"required,min=1,max=256"`
	}) {
		c.JSON(http.StatusOK, service.FetchPluginInstallationTasks(request.TenantID, request.Page, request.PageSize))
	})
}

func FetchPluginInstallationTask(c *gin.Context) {
	BindRequest(c, func(request struct {
		TenantID string `uri:"tenant_id" validate:"required"`
		TaskID   string `uri:"id" validate:"required"`
	}) {
		c.JSON(http.StatusOK, service.FetchPluginInstallationTask(request.TenantID, request.TaskID))
	})
}

func DeletePluginInstallationTask(c *gin.Context) {
	BindRequest(c, func(request struct {
		TenantID string `uri:"tenant_id" validate:"required"`
		TaskID   string `uri:"id" validate:"required"`
	}) {
		c.JSON(http.StatusOK, service.DeletePluginInstallationTask(request.TenantID, request.TaskID))
	})
}

func DeletePluginInstallationItemFromTask(c *gin.Context) {
	BindRequest(c, func(request struct {
		TenantID   string                                 `uri:"tenant_id" validate:"required"`
		TaskID     string                                 `uri:"id" validate:"required"`
		Identifier plugin_entities.PluginUniqueIdentifier `uri:"identifier" validate:"required,plugin_unique_identifier"`
	}) {
		c.JSON(http.StatusOK, service.DeletePluginInstallationItemFromTask(request.TenantID, request.TaskID, request.Identifier))
	})
}

func FetchPluginManifest(c *gin.Context) {
	BindRequest(c, func(request struct {
		TenantID               string                                 `uri:"tenant_id" validate:"required"`
		PluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier `form:"plugin_unique_identifier" validate:"required,plugin_unique_identifier"`
	}) {
		c.JSON(http.StatusOK, service.FetchPluginManifest(request.TenantID, request.PluginUniqueIdentifier))
	})
}

func UninstallPlugin(c *gin.Context) {
	BindRequest(c, func(request struct {
		TenantID             string `uri:"tenant_id" validate:"required"`
		PluginInstallationID string `json:"plugin_installation_id" validate:"required"`
	}) {
		c.JSON(http.StatusOK, service.UninstallPlugin(request.TenantID, request.PluginInstallationID))
	})
}

func FetchPluginFromIdentifier(c *gin.Context) {
	BindRequest(c, func(request struct {
		PluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier `form:"plugin_unique_identifier" validate:"required,plugin_unique_identifier"`
	}) {
		c.JSON(http.StatusOK, service.FetchPluginFromIdentifier(request.PluginUniqueIdentifier))
	})
}

func ListPlugins(c *gin.Context) {
	BindRequest(c, func(request struct {
		TenantID string `uri:"tenant_id" validate:"required"`
		Page     int    `form:"page" validate:"required,min=1"`
		PageSize int    `form:"page_size" validate:"required,min=1,max=256"`
	}) {
		c.JSON(http.StatusOK, service.ListPlugins(request.TenantID, request.Page, request.PageSize))
	})
}

func BatchFetchPluginInstallationByIDs(c *gin.Context) {
	BindRequest(c, func(request struct {
		TenantID  string   `uri:"tenant_id" validate:"required"`
		PluginIDs []string `json:"plugin_ids" validate:"required,max=256"`
	}) {
		c.JSON(http.StatusOK, service.BatchFetchPluginInstallationByIDs(request.TenantID, request.PluginIDs))
	})
}

func FetchMissingPluginInstallations(c *gin.Context) {
	BindRequest(c, func(request struct {
		TenantID                string                                   `uri:"tenant_id" validate:"required"`
		PluginUniqueIdentifiers []plugin_entities.PluginUniqueIdentifier `json:"plugin_unique_identifiers" validate:"required,max=256,dive,plugin_unique_identifier"`
	}) {
		c.JSON(http.StatusOK, service.FetchMissingPluginInstallations(request.TenantID, request.PluginUniqueIdentifiers))
	})
}
