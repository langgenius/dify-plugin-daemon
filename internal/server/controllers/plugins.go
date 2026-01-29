package controllers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/service"
	"github.com/langgenius/dify-plugin-daemon/internal/types/app"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
)

func GetAsset(c *gin.Context) {
	pluginManager := plugin_manager.Manager()
	asset, err := pluginManager.GetAsset(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusInternalServerError, exception.InternalServerError(err).ToResponse())
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", asset)
}

func UploadPlugin(app *app.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		difyPkgFileHeader, err := c.FormFile("dify_pkg")
		if err != nil {
			c.JSON(http.StatusOK, exception.BadRequestError(err).ToResponse())
			return
		}

		tenantId := c.Param("tenant_id")
		if tenantId == "" {
			c.JSON(http.StatusOK, exception.BadRequestError(errors.New("tenant ID is required")).ToResponse())
			return
		}

		if difyPkgFileHeader.Size > app.MaxPluginPackageSize {
			c.JSON(http.StatusOK, exception.BadRequestError(errors.New("file size exceeds the maximum limit")).ToResponse())
			return
		}

		verifySignature := c.PostForm("verify_signature") == "true"

		difyPkgFile, err := difyPkgFileHeader.Open()
		if err != nil {
			c.JSON(http.StatusOK, exception.BadRequestError(err).ToResponse())
			return
		}
		defer difyPkgFile.Close()

		c.JSON(http.StatusOK, service.UploadPluginPkg(app, c, tenantId, difyPkgFile, verifySignature))
	}
}

func UploadBundle(app *app.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		difyBundleFileHeader, err := c.FormFile("dify_bundle")
		if err != nil {
			c.JSON(http.StatusOK, exception.BadRequestError(err).ToResponse())
			return
		}

		tenantId := c.Param("tenant_id")
		if tenantId == "" {
			c.JSON(http.StatusOK, exception.BadRequestError(errors.New("tenant ID is required")).ToResponse())
			return
		}

		if difyBundleFileHeader.Size > app.MaxBundlePackageSize {
			c.JSON(http.StatusOK, exception.BadRequestError(errors.New("file size exceeds the maximum limit")).ToResponse())
			return
		}

		verifySignature := c.PostForm("verify_signature") == "true"

		difyBundleFile, err := difyBundleFileHeader.Open()
		if err != nil {
			c.JSON(http.StatusOK, exception.BadRequestError(err).ToResponse())
			return
		}
		defer difyBundleFile.Close()

		c.JSON(http.StatusOK, service.UploadPluginBundle(app, c, tenantId, difyBundleFile, verifySignature))
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
			if request.OriginalPluginUniqueIdentifier == request.NewPluginUniqueIdentifier {
				c.JSON(http.StatusOK, exception.BadRequestError(errors.New("original and new plugin unique identifier are the same")).ToResponse())
				return
			}

			if request.OriginalPluginUniqueIdentifier.PluginID() != request.NewPluginUniqueIdentifier.PluginID() {
				c.JSON(http.StatusOK, exception.BadRequestError(errors.New("original and new plugin id are different")).ToResponse())
				return
			}

			c.JSON(http.StatusOK, service.UpgradePlugin(
				c.Request.Context(),
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
			Metas                   []map[string]any                         `json:"metas" validate:"omitempty"`
		}) {
			if request.Metas == nil {
				request.Metas = []map[string]any{}
			}

			if len(request.Metas) != len(request.PluginUniqueIdentifiers) {
				c.JSON(http.StatusOK, exception.BadRequestError(errors.New("the number of metas must be equal to the number of plugin unique identifiers")).ToResponse())
				return
			}

			for i := range request.Metas {
				if request.Metas[i] == nil {
					request.Metas[i] = map[string]any{}
				}
			}

			c.JSON(http.StatusOK, service.InstallMultiplePluginsToTenant(
				c.Request.Context(), app, request.TenantID, request.PluginUniqueIdentifiers, request.Source, request.Metas,
			))
		})
	}
}

func ReinstallPluginFromIdentifier(app *app.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		BindRequest(c, func(request struct {
			PluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier `json:"plugin_unique_identifier" validate:"required,plugin_unique_identifier"`
		}) {
			service.ReinstallPluginFromIdentifier(c, app, request.PluginUniqueIdentifier)
		})
	}
}

func DecodePluginFromIdentifier(app *app.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		BindRequest(c, func(request struct {
			PluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier `form:"plugin_unique_identifier" validate:"required,plugin_unique_identifier"`
		}) {
			c.JSON(http.StatusOK, service.DecodePluginFromIdentifier(app, request.PluginUniqueIdentifier))
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

func DeleteAllPluginInstallationTasks(c *gin.Context) {
	BindRequest(c, func(request struct {
		TenantID string `uri:"tenant_id" validate:"required"`
	}) {
		c.JSON(http.StatusOK, service.DeleteAllPluginInstallationTasks(request.TenantID))
	})
}

func DeletePluginInstallationItemFromTask(c *gin.Context) {
	BindRequest(c, func(request struct {
		TenantID   string `uri:"tenant_id" validate:"required"`
		TaskID     string `uri:"id" validate:"required"`
		Identifier string `uri:"identifier" validate:"required"`
	}) {
		identifierString := strings.TrimLeft(request.Identifier, "/")
		identifier, err := plugin_entities.NewPluginUniqueIdentifier(identifierString)
		if err != nil {
			c.JSON(http.StatusOK, exception.BadRequestError(err).ToResponse())
			return
		}

		c.JSON(http.StatusOK, service.DeletePluginInstallationItemFromTask(request.TenantID, request.TaskID, identifier))
	})
}

func FetchPluginManifest(c *gin.Context) {
	BindRequest(c, func(request struct {
		TenantID               string                                 `uri:"tenant_id" validate:"required"`
		PluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier `form:"plugin_unique_identifier" validate:"required,plugin_unique_identifier"`
	}) {
		c.JSON(http.StatusOK, service.FetchPluginManifest(request.PluginUniqueIdentifier))
	})
}

func FetchPluginReadme(c *gin.Context) {
	BindRequest(c, func(request struct {
		PluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier `form:"plugin_unique_identifier" validate:"required,plugin_unique_identifier"`
		Language               string                                 `form:"language" validate:"omitempty"`
	}) {
		c.JSON(http.StatusOK, service.FetchPluginReadme(request.PluginUniqueIdentifier, request.Language))
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

func ExtractPluginAsset(c *gin.Context) {
	BindRequest(c, func(request struct {
		TenantID               string                                 `uri:"tenant_id" validate:"required"`
		PluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier "form:\"plugin_unique_identifier\" validate:\"required,plugin_unique_identifier\""
		FilePath               string                                 `form:"file_path" validate:"required"`
	}) {
		manager := plugin_manager.Manager()
		asset, err := manager.ExtractPluginAsset(request.PluginUniqueIdentifier, request.FilePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, exception.InternalServerError(err).ToResponse())
			return
		}
		c.Data(http.StatusOK, "application/octet-stream", asset)
	})
}

func SwitchServerlessEndpoint(c *gin.Context) {
	BindRequest(c, func(request struct {
		PluginUniqueIdentifier plugin_entities.PluginUniqueIdentifier `json:"plugin_unique_identifier" validate:"required,plugin_unique_identifier"`
		FunctionName           string                                 `json:"function_name" validate:"required"`
		FunctionURL            string                                 `json:"function_url" validate:"required"`
	}) {
		c.JSON(
			http.StatusOK,
			service.SwitchServerlessEndpoint(
				request.PluginUniqueIdentifier,
				request.FunctionName,
				request.FunctionURL,
			),
		)
	})
}
