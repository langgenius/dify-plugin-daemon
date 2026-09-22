package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/server/contracts"
	"github.com/langgenius/dify-plugin-daemon/internal/service"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
	"github.com/langgenius/dify-plugin-daemon/internal/types/models"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/stretchr/testify/require"
)

const toolManagementTestPath = "/plugin/tenant-1/management"

func TestToolManagementSuccessPreservesLegacyPayload(t *testing.T) {
	for _, operation := range []string{"list", "get"} {
		t.Run(operation, func(t *testing.T) {
			tool := toolManagementFixture(t)
			var called bool
			handlers := toolManagementHandlers{
				listTools: func(tenantID string, page, pageSize int) ([]service.InstalledTool, exception.PluginDaemonError) {
					called = true
					require.Equal(t, "tenant-1", tenantID)
					require.Equal(t, 1, page)
					require.Equal(t, 256, pageSize)
					return []service.InstalledTool{tool}, nil
				},
				getTool: func(tenantID, pluginID, provider string) (service.InstalledTool, exception.PluginDaemonError) {
					called = true
					require.Equal(t, "tenant-1", tenantID)
					require.Equal(t, "org/plugin", pluginID)
					require.Equal(t, "provider", provider)
					return tool, nil
				},
			}
			path := toolManagementTestPath + "/tool?plugin_id=org%2Fplugin&provider=provider"
			var payload any = toolManagementLegacyTool(tool)
			if operation == "list" {
				path = toolManagementTestPath + "/tools?page=1&page_size=256"
				payload = []any{payload}
			}

			response := toolManagementRequest(t, handlers, path)
			require.True(t, called)
			require.Equal(t, http.StatusOK, response.Code)
			toolManagementRequireJSON(t, entities.NewSuccessResponse(payload), response.Body.Bytes())
		})
	}
}

func TestToolManagementEmptyListRemainsArray(t *testing.T) {
	for name, providers := range map[string][]service.InstalledTool{"nil": nil, "empty": {}} {
		t.Run(name, func(t *testing.T) {
			response := toolManagementRequest(t, toolManagementHandlers{
				listTools: func(string, int, int) ([]service.InstalledTool, exception.PluginDaemonError) {
					return providers, nil
				},
			}, toolManagementTestPath+"/tools?page=1&page_size=1")
			require.Equal(t, http.StatusOK, response.Code)
			require.JSONEq(t, `{"code":0,"message":"success","data":[]}`, response.Body.String())
		})
	}
}

func TestToolManagementNilDeclarationPreservesLegacyPayload(t *testing.T) {
	tool := toolManagementFixture(t)
	tool.Declaration = nil
	response := toolManagementGetFixture(t, tool)
	toolManagementRequireJSON(t, entities.NewSuccessResponse(toolManagementLegacyTool(tool)), response.Body.Bytes())
}

func TestToolManagementNullableCollectionsPreserveLegacyPayload(t *testing.T) {
	tests := []struct {
		name   string
		change func(*plugin_entities.ToolProviderDeclaration)
	}{
		{
			name: "nil provider collections normalize to empty arrays",
			change: func(declaration *plugin_entities.ToolProviderDeclaration) {
				declaration.Tools = nil
				declaration.CredentialsSchema = nil
				declaration.OAuthSchema = nil
				declaration.Identity.Tags = nil
			},
		},
		{
			name: "nil parameters and omitted output schema",
			change: func(declaration *plugin_entities.ToolProviderDeclaration) {
				declaration.Tools[0].Parameters = nil
				declaration.Tools[0].OutputSchema = nil
			},
		},
		{
			name: "empty parameters and omitted empty output schema",
			change: func(declaration *plugin_entities.ToolProviderDeclaration) {
				declaration.Tools[0].Parameters = []plugin_entities.ToolParameter{}
				declaration.Tools[0].OutputSchema = plugin_entities.ToolOutputSchema{}
			},
		},
		{
			name: "nullable parameter fields and absent reset targets",
			change: func(declaration *plugin_entities.ToolProviderDeclaration) {
				parameter := &declaration.Tools[0].Parameters[0]
				parameter.Scope = nil
				parameter.AutoGenerate = nil
				parameter.Template = nil
				parameter.Default = nil
				parameter.Min = nil
				parameter.Max = nil
				parameter.Precision = nil
				parameter.Options = nil
				parameter.ResetOnChange = nil
			},
		},
		{
			name: "empty parameter options and reset targets",
			change: func(declaration *plugin_entities.ToolProviderDeclaration) {
				parameter := &declaration.Tools[0].Parameters[0]
				parameter.Options = []plugin_entities.ParameterOption{}
				parameter.ResetOnChange = []string{}
			},
		},
		{
			name: "nullable OAuth collections",
			change: func(declaration *plugin_entities.ToolProviderDeclaration) {
				declaration.OAuthSchema = &plugin_entities.OAuthSchema{}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tool := toolManagementFixture(t)
			test.change(tool.Declaration)
			response := toolManagementGetFixture(t, tool)
			toolManagementRequireJSON(t, entities.NewSuccessResponse(toolManagementLegacyTool(tool)), response.Body.Bytes())
		})
	}
}

func TestToolManagementParameterDefaultsPreserveOpenJSON(t *testing.T) {
	tests := []struct {
		name  string
		value any
	}{
		{name: "null"},
		{name: "true", value: true},
		{name: "false", value: false},
		{name: "zero", value: 0},
		{name: "fraction", value: 1.25},
		{name: "empty string", value: ""},
		{name: "string", value: "value"},
		{name: "empty array", value: []any{}},
		{name: "empty object", value: map[string]any{}},
		{name: "nested JSON", value: map[string]any{"items": []any{nil, false, 0, "", map[string]any{"enabled": true}}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tool := toolManagementFixture(t)
			parameter := &tool.Declaration.Tools[0].Parameters[0]
			parameter.Type = plugin_entities.TOOL_PARAMETER_TYPE_ANY
			parameter.Multiple = false
			parameter.Default = test.value
			response := toolManagementGetFixture(t, tool)
			toolManagementRequireJSON(t, entities.NewSuccessResponse(toolManagementLegacyTool(tool)), response.Body.Bytes())
		})
	}
}

func TestToolManagementParameterBoundsPreserveFloat64Precision(t *testing.T) {
	tool := toolManagementFixture(t)
	minimum, maximum := 0.1234567890123456, 123456789.12345679
	parameter := &tool.Declaration.Tools[0].Parameters[0]
	parameter.Min = &minimum
	parameter.Max = &maximum
	response := toolManagementGetFixture(t, tool)
	toolManagementRequireJSON(t, entities.NewSuccessResponse(toolManagementLegacyTool(tool)), response.Body.Bytes())

	var payload map[string]any
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
	declaration := payload["data"].(map[string]any)["declaration"].(map[string]any)
	parameters := declaration["tools"].([]any)[0].(map[string]any)["parameters"].([]any)
	require.Equal(t, minimum, parameters[0].(map[string]any)["min"])
	require.Equal(t, maximum, parameters[0].(map[string]any)["max"])
}

func TestToolManagementResponseValidationRejectsInvalidPayloads(t *testing.T) {
	path := toolManagementTestPath + "/tool?plugin_id=org%2Fplugin&provider=provider"
	response := toolManagementGetFixture(t, toolManagementFixture(t))
	tests := []struct {
		name   string
		change func(map[string]any)
	}{
		{
			name: "invalid parameter enum",
			change: func(payload map[string]any) {
				declaration := payload["data"].(map[string]any)["declaration"].(map[string]any)
				parameters := declaration["tools"].([]any)[0].(map[string]any)["parameters"].([]any)
				parameters[0].(map[string]any)["type"] = "unsupported-parameter-type"
			},
		},
		{
			name: "missing required installation field",
			change: func(payload map[string]any) {
				delete(payload["data"].(map[string]any), "plugin_id")
			},
		},
		{
			name: "unknown closed object field",
			change: func(payload map[string]any) {
				payload["data"].(map[string]any)["unknown_field"] = "unexpected"
			},
		},
		{
			name: "object credential default",
			change: func(payload map[string]any) {
				declaration := payload["data"].(map[string]any)["declaration"].(map[string]any)
				credentials := declaration["credentials_schema"].([]any)
				credentials[0].(map[string]any)["default"] = map[string]any{"nested": true}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var payload map[string]any
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
			test.change(payload)
			body, err := json.Marshal(payload)
			require.NoError(t, err)
			request := httptest.NewRequest(http.MethodGet, path, nil)
			require.Error(t, toolManagementValidateResponse(t, request, response.Code, response.Header(), body))
		})
	}

	t.Run("non-null error data", func(t *testing.T) {
		errorResponse := toolManagementRequest(t, toolManagementHandlers{
			getTool: func(string, string, string) (service.InstalledTool, exception.PluginDaemonError) {
				return service.InstalledTool{}, exception.ErrPluginNotFound()
			},
		}, path)
		var payload map[string]any
		require.NoError(t, json.Unmarshal(errorResponse.Body.Bytes(), &payload))
		payload["data"] = map[string]any{}
		body, err := json.Marshal(payload)
		require.NoError(t, err)
		request := httptest.NewRequest(http.MethodGet, path, nil)
		require.Error(t, toolManagementValidateResponse(t, request, errorResponse.Code, errorResponse.Header(), body))
	})

	t.Run("undeclared HTTP status", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		require.Error(t, toolManagementValidateResponse(t, request, http.StatusTeapot, response.Header(), response.Body.Bytes()))
	})
}

func TestToolManagementNonJSONDefaultReturnsCompleteHTTP500Envelope(t *testing.T) {
	for _, operation := range []string{"list", "get"} {
		t.Run(operation, func(t *testing.T) {
			tool := toolManagementFixture(t)
			tool.Declaration.Tools[0].Parameters[0].Default = func() {}
			handlers := toolManagementHandlers{
				listTools: func(string, int, int) ([]service.InstalledTool, exception.PluginDaemonError) {
					return []service.InstalledTool{toolManagementFixture(t), tool}, nil
				},
				getTool: func(string, string, string) (service.InstalledTool, exception.PluginDaemonError) {
					return tool, nil
				},
			}
			path := toolManagementTestPath + "/tool?plugin_id=org%2Fplugin&provider=provider"
			if operation == "list" {
				path = toolManagementTestPath + "/tools?page=1&page_size=256"
			}
			response := toolManagementRequest(t, handlers, path)
			require.Equal(t, http.StatusInternalServerError, response.Code)
			var envelope entities.Response
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &envelope))
			require.Equal(t, -500, envelope.Code)
			require.Nil(t, envelope.Data)
			var detail map[string]any
			require.NoError(t, json.Unmarshal([]byte(envelope.Message), &detail))
			require.Equal(t, exception.PluginDaemonInternalServerError, detail["error_type"])
			require.Contains(t, detail["message"], "unsupported type")
		})
	}
}

func TestToolManagementServiceErrorsPreserveHTTP200Envelope(t *testing.T) {
	errors := []exception.PluginDaemonError{
		exception.ErrPluginNotFound(),
		exception.ErrorWithTypeAndCode("database unavailable", exception.PluginDaemonInternalServerError, -500),
	}
	for _, serviceError := range errors {
		for _, operation := range []string{"list", "get"} {
			t.Run(operation+"/"+serviceError.Error(), func(t *testing.T) {
				handlers := toolManagementHandlers{
					listTools: func(string, int, int) ([]service.InstalledTool, exception.PluginDaemonError) {
						return nil, serviceError
					},
					getTool: func(string, string, string) (service.InstalledTool, exception.PluginDaemonError) {
						return service.InstalledTool{}, serviceError
					},
				}
				path := toolManagementTestPath + "/tool?plugin_id=org%2Fplugin&provider=provider"
				if operation == "list" {
					path = toolManagementTestPath + "/tools?page=1&page_size=256"
				}
				response := toolManagementRequest(t, handlers, path)
				require.Equal(t, http.StatusOK, response.Code)
				toolManagementRequireJSON(t, serviceError.ToResponse(), response.Body.Bytes())
			})
		}
	}
}

func TestToolManagementInvalidRequestsDoNotCallService(t *testing.T) {
	paths := []string{
		"/tools",
		"/tools?page=1",
		"/tools?page_size=1",
		"/tools?page=0&page_size=1",
		"/tools?page=-1&page_size=1",
		"/tools?page=1&page_size=0",
		"/tools?page=1&page_size=257",
		"/tools?page=invalid&page_size=1",
		"/tools?page=1&page_size=invalid",
		"/tools?page=1&page=2&page_size=1",
		"/tools?page=1&page_size=1&page_size=2",
		"/tool",
		"/tool?plugin_id=org%2Fplugin",
		"/tool?provider=provider",
		"/tool?plugin_id=&provider=provider",
		"/tool?plugin_id=org%2Fplugin&provider=",
		"/tool?plugin_id=org%2Fplugin&provider=provider&provider=other",
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			calls := 0
			handlers := toolManagementHandlers{
				listTools: func(string, int, int) ([]service.InstalledTool, exception.PluginDaemonError) {
					calls++
					return nil, nil
				},
				getTool: func(string, string, string) (service.InstalledTool, exception.PluginDaemonError) {
					calls++
					return service.InstalledTool{}, nil
				},
			}
			response := toolManagementRequest(t, handlers, toolManagementTestPath+path)
			require.Equal(t, http.StatusBadRequest, response.Code)
			require.Zero(t, calls)
			var envelope entities.Response
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &envelope))
			require.Equal(t, -400, envelope.Code)
			require.Nil(t, envelope.Data)
			var detail map[string]any
			require.NoError(t, json.Unmarshal([]byte(envelope.Message), &detail))
			require.Equal(t, exception.PluginDaemonBadRequestError, detail["error_type"])
		})
	}
}

func TestToolManagementAdditionalQueryParametersRemainAccepted(t *testing.T) {
	tool := toolManagementFixture(t)
	calls := 0
	handlers := toolManagementHandlers{
		listTools: func(string, int, int) ([]service.InstalledTool, exception.PluginDaemonError) {
			calls++
			return []service.InstalledTool{}, nil
		},
		getTool: func(string, string, string) (service.InstalledTool, exception.PluginDaemonError) {
			calls++
			return tool, nil
		},
	}
	for _, path := range []string{
		"/tools?page=1&page_size=1&unused=value",
		"/tool?plugin_id=org%2Fplugin&provider=provider&unused=value",
	} {
		response := toolManagementRequest(t, handlers, toolManagementTestPath+path)
		require.Equal(t, http.StatusOK, response.Code)
	}
	require.Equal(t, 2, calls)
}

func toolManagementRequest(t *testing.T, handlers toolManagementHandlers, path string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	require.NoError(t, registerToolManagementRoutes(router, handlers))
	request := httptest.NewRequest(http.MethodGet, path, nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.NoError(t, toolManagementValidateResponse(t, request, response.Code, response.Header(), response.Body.Bytes()))
	return response
}

func toolManagementValidateResponse(t *testing.T, request *http.Request, status int, header http.Header, body []byte) error {
	t.Helper()
	spec, err := contracts.GetSpec()
	require.NoError(t, err)
	specRouter, err := legacy.NewRouter(spec)
	require.NoError(t, err)
	route, pathParams, err := specRouter.FindRoute(request)
	require.NoError(t, err)
	validation := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request: request, PathParams: pathParams, Route: route,
		},
		Status: status,
		Header: header,
		Options: &openapi3filter.Options{
			IncludeResponseStatus: true,
		},
	}
	return openapi3filter.ValidateResponse(context.Background(), validation.SetBodyBytes(body))
}

func toolManagementGetFixture(t *testing.T, tool service.InstalledTool) *httptest.ResponseRecorder {
	t.Helper()
	response := toolManagementRequest(t, toolManagementHandlers{
		getTool: func(string, string, string) (service.InstalledTool, exception.PluginDaemonError) {
			return tool, nil
		},
	}, toolManagementTestPath+"/tool?plugin_id=org%2Fplugin&provider=provider")
	require.Equal(t, http.StatusOK, response.Code)
	return response
}

func toolManagementLegacyTool(tool service.InstalledTool) any {
	return struct {
		models.ToolInstallation
		Declaration *plugin_entities.ToolProviderDeclaration `json:"declaration"`
	}{
		ToolInstallation: models.ToolInstallation{
			Model:    models.Model{ID: tool.ID, CreatedAt: tool.CreatedAt, UpdatedAt: tool.UpdatedAt},
			TenantID: tool.TenantID, Provider: tool.Provider,
			PluginUniqueIdentifier: tool.PluginUniqueIdentifier, PluginID: tool.PluginID,
		},
		Declaration: tool.Declaration,
	}
}

func toolManagementRequireJSON(t *testing.T, expected any, actual []byte) {
	t.Helper()
	encoded, err := json.Marshal(expected)
	require.NoError(t, err)
	require.JSONEq(t, string(encoded), string(actual))
}

func toolManagementFixture(t *testing.T) service.InstalledTool {
	t.Helper()
	var declaration plugin_entities.ToolProviderDeclaration
	require.NoError(t, json.Unmarshal([]byte(`{
		"identity": {
			"author":"author", "name":"provider", "label":{"en_US":"Provider","zh_Hans":"Provider zh"},
			"description":{"en_US":"Provider description"}, "icon":"icon.svg", "icon_dark":"dark.svg", "tags":["search"]
		},
		"credentials_schema":[{
			"name":"account", "type":"text-input", "scope":null, "required":false, "default":"",
			"options":null, "multiple":false, "label":{"en_US":"Account"},
			"help":{"en_US":"Account help"}, "url":"https://example.com/help",
			"placeholder":{"en_US":"Account name"}, "reset_on_change":["database"]
		}],
		"oauth_schema":{
			"client_schema":[{"name":"client_id","type":"text-input","label":{"en_US":"Client ID"},"default":false}],
			"credentials_schema":[{"name":"access_token","type":"secret-input","label":{"en_US":"Access token"}}]
		},
		"tools":[{
			"identity":{"author":"author","name":"search","label":{"en_US":"Search"}},
			"description":{"human":{"en_US":"Search description"},"llm":"Search records"},
			"has_runtime_parameters":true,
			"parameters":[{
				"name":"database", "label":{"en_US":"Database"}, "human_description":{"en_US":"Choose a database"},
				"type":"dynamic-select", "form":"form", "scope":"object", "llm_description":"Database selection",
				"required":true, "auto_generate":{"type":"prompt_instruction"}, "template":{"enabled":false},
				"default":["first"], "min":0, "max":2.5, "multiple":true, "precision":0,
				"options":[{"value":"first","label":{"en_US":"First"},"icon":"option.svg"}],
				"reset_on_change":["table"]
			}],
			"output_schema":{
				"type":"object", "additionalProperties":true,
				"properties":{"records":{"type":"array","items":{"anyOf":[{"type":"string"},{"type":"null"}]}}},
				"x-values":[null,false,0,"",{},[]]
			}
		}]
	}`), &declaration))
	return service.InstalledTool{
		ID:        "0197ed41-854c-7000-8000-000000000001",
		CreatedAt: time.Date(2026, 9, 22, 12, 0, 0, 123456789, time.FixedZone("fixture", 8*60*60)),
		UpdatedAt: time.Date(2026, 9, 22, 5, 0, 0, 0, time.UTC),
		TenantID:  "tenant-1", Provider: "provider", PluginID: "org/plugin",
		PluginUniqueIdentifier: "org/plugin:1.0.0@checksum", Declaration: &declaration,
	}
}
