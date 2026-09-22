package controllers

import (
	"context"
	"errors"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gin-gonic/gin"
	ginmiddleware "github.com/oapi-codegen/gin-middleware"

	"github.com/langgenius/dify-plugin-daemon/internal/server/contracts"
	"github.com/langgenius/dify-plugin-daemon/internal/service"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
)

type toolManagementHandlers struct {
	listTools func(string, int, int) ([]service.InstalledTool, exception.PluginDaemonError)
	getTool   func(string, string, string) (service.InstalledTool, exception.PluginDaemonError)
}

func RegisterToolManagementRoutes(router gin.IRouter) error {
	return registerToolManagementRoutes(router, toolManagementHandlers{
		listTools: service.ListTools,
		getTool:   service.GetTool,
	})
}

func registerToolManagementRoutes(router gin.IRouter, handlers contracts.StrictServerInterface) error {
	spec, err := contracts.GetSpec()
	if err != nil {
		return err
	}
	if err := spec.Validate(context.Background()); err != nil {
		return err
	}
	badRequest := func(c *gin.Context, err error) {
		c.AbortWithStatusJSON(http.StatusBadRequest, toolManagementError(exception.BadRequestError(err)))
	}
	internalError := func(c *gin.Context, err error) {
		c.AbortWithStatusJSON(http.StatusInternalServerError, toolManagementError(exception.InternalServerError(err)))
	}
	validator := ginmiddleware.OapiRequestValidatorWithOptions(spec, &ginmiddleware.Options{
		Options: openapi3filter.Options{
			// CheckingKey authenticates this router group before contract validation.
			AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
		},
		ErrorHandler: func(c *gin.Context, message string, _ int) {
			badRequest(c, errors.New(message))
		},
	})
	strict := contracts.NewStrictHandlerWithOptions(handlers, nil, contracts.StrictGinServerOptions{
		RequestErrorHandlerFunc:  badRequest,
		HandlerErrorFunc:         internalError,
		ResponseErrorHandlerFunc: internalError,
	})
	contracts.RegisterHandlersWithOptions(router, strict, contracts.GinServerOptions{
		Middlewares: []contracts.MiddlewareFunc{contracts.MiddlewareFunc(validator)},
		ErrorHandler: func(c *gin.Context, err error, _ int) {
			badRequest(c, err)
		},
	})
	return nil
}

func toolManagementError(err exception.PluginDaemonError) contracts.DaemonErrorResponse {
	response := err.ToResponse()
	return contracts.DaemonErrorResponse{Code: response.Code, Message: response.Message}
}

func (h toolManagementHandlers) ListTools(_ context.Context, request contracts.ListToolsRequestObject) (contracts.ListToolsResponseObject, error) {
	tools, daemonErr := h.listTools(request.TenantID, request.Params.Page, request.Params.PageSize)
	var response contracts.ListToolsResponse
	if daemonErr != nil {
		if err := response.FromDaemonErrorResponse(toolManagementError(daemonErr)); err != nil {
			return nil, err
		}
	} else {
		data := make([]contracts.ToolProviderInstallation, 0, len(tools))
		for _, tool := range tools {
			data = append(data, toolProviderInstallation(tool))
		}
		if err := response.FromListToolsSuccessResponse(contracts.ListToolsSuccessResponse{
			Code: 0, Message: "success", Data: data,
		}); err != nil {
			return nil, err
		}
	}
	return contracts.ListTools200JSONResponse(response), nil
}

func (h toolManagementHandlers) GetTool(_ context.Context, request contracts.GetToolRequestObject) (contracts.GetToolResponseObject, error) {
	tool, daemonErr := h.getTool(request.TenantID, request.Params.PluginID, request.Params.Provider)
	var response contracts.GetToolResponse
	if daemonErr != nil {
		if err := response.FromDaemonErrorResponse(toolManagementError(daemonErr)); err != nil {
			return nil, err
		}
	} else {
		if err := response.FromGetToolSuccessResponse(contracts.GetToolSuccessResponse{
			Code: 0, Message: "success", Data: toolProviderInstallation(tool),
		}); err != nil {
			return nil, err
		}
	}
	return contracts.GetTool200JSONResponse(response), nil
}

func CheckToolExistence(c *gin.Context) {
	BindRequest(c, func(request struct {
		TenantID    string                              `uri:"tenant_id" validate:"required"`
		ProviderIDS []service.RequestCheckToolExistence `json:"provider_ids" validate:"required,dive"`
	}) {
		c.JSON(http.StatusOK, service.CheckToolExistence(request.TenantID, request.ProviderIDS))
	})
}
