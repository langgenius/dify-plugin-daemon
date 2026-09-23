package controllers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
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
			// Validate bound query values below: kin-openapi parses integer strings with base 0.
			ExcludeRequestQueryParams: true,
		},
		ErrorHandler: func(c *gin.Context, message string, _ int) {
			badRequest(c, errors.New(message))
		},
	})
	strict := contracts.NewStrictHandlerWithOptions(handlers, []contracts.StrictMiddlewareFunc{
		toolManagementQueryValidator(spec, badRequest),
	}, contracts.StrictGinServerOptions{
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

func toolManagementQueryValidator(spec *openapi3.T, badRequest func(*gin.Context, error)) contracts.StrictMiddlewareFunc {
	parameters := make(map[string]openapi3.Parameters)
	for _, path := range spec.Paths.Map() {
		for _, operation := range path.Operations() {
			parameters[operation.OperationID] = operation.Parameters
		}
	}
	return func(next contracts.StrictHandlerFunc, operationID string) contracts.StrictHandlerFunc {
		return func(c *gin.Context, request any) (any, error) {
			var values map[string]any
			switch request := request.(type) {
			case contracts.ListToolsRequestObject:
				values = map[string]any{"page": request.Params.Page, "page_size": request.Params.PageSize}
			case contracts.GetToolRequestObject:
				values = map[string]any{"plugin_id": request.Params.PluginID, "provider": request.Params.Provider}
			default:
				return nil, fmt.Errorf("unsupported tool management request: %T", request)
			}
			for _, parameter := range parameters[operationID] {
				if parameter.Value.In != openapi3.ParameterInQuery {
					continue
				}
				value, ok := values[parameter.Value.Name]
				if !ok {
					return nil, fmt.Errorf("missing bound query parameter: %s", parameter.Value.Name)
				}
				if err := parameter.Value.Schema.Value.VisitJSON(value, openapi3.EnableJSONSchema2020()); err != nil {
					badRequest(c, fmt.Errorf("invalid query parameter %s: %w", parameter.Value.Name, err))
					return nil, nil
				}
			}
			return next(c, request)
		}
	}
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
