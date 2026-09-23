package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/langgenius/dify-plugin-daemon/internal/server/contracts"
	"github.com/langgenius/dify-plugin-daemon/internal/server/controllers"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
)

func TestToolManagementAuthenticationPrecedesValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	require.NoError(t, controllers.RegisterToolManagementRoutes(router.Group("", CheckingKey("test-key"))))
	spec, err := contracts.GetSpec()
	require.NoError(t, err)
	schemaRouter, err := legacy.NewRouter(spec)
	require.NoError(t, err)
	for _, path := range []string{"/tools", "/tool"} {
		for _, test := range []struct {
			name       string
			key        string
			statusCode int
		}{
			{name: "missing key", statusCode: http.StatusUnauthorized},
			{name: "wrong key", key: "wrong-key", statusCode: http.StatusUnauthorized},
			{name: "valid key and invalid parameters", key: "test-key", statusCode: http.StatusBadRequest},
		} {
			t.Run(path+"/"+test.name, func(t *testing.T) {
				request := httptest.NewRequest(http.MethodGet, "/plugin/tenant-1/management"+path, nil)
				request.Header.Set("X-Api-Key", test.key)
				response := httptest.NewRecorder()
				router.ServeHTTP(response, request)
				require.Equal(t, test.statusCode, response.Code)

				route, pathParams, err := schemaRouter.FindRoute(request)
				require.NoError(t, err)
				validation := &openapi3filter.ResponseValidationInput{
					RequestValidationInput: &openapi3filter.RequestValidationInput{
						Request: request, PathParams: pathParams, Route: route,
					},
					Status: response.Code,
					Header: response.Header(),
					Options: &openapi3filter.Options{
						IncludeResponseStatus: true,
					},
				}
				validation.SetBodyBytes(response.Body.Bytes())
				require.NoError(t, openapi3filter.ValidateResponse(context.Background(), validation))
				if test.statusCode == http.StatusUnauthorized {
					expected, err := json.Marshal(exception.UnauthorizedError().ToResponse())
					require.NoError(t, err)
					require.JSONEq(t, string(expected), response.Body.String())
				}
			})
		}
	}
}
