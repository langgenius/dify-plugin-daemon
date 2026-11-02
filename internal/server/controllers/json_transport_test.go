package controllers

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
	"github.com/stretchr/testify/require"
)

func checkDispatchJSON[T any](t *testing.T, payload string) {
	t.Helper()
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/", strings.NewReader(payload))
	c.Request.Header.Set("Content-Type", "application/json")
	called := false
	BindRequest(c, func(request plugin_entities.InvokePluginRequest[T]) {
		called = true
		encoded, err := json.Marshal(parser.StructToMap(request.Data))
		require.NoError(t, err)
		require.Contains(t, string(encoded), `"n":9007199254740993`)
		require.Contains(t, string(encoded), `"state":[false,0,{},[]]`)
	})
	require.True(t, called)
}

func TestDispatchPreservesToolArguments(t *testing.T) {
	checkDispatchJSON[requests.RequestInvokeTool](t, `{"tenant_id":"t","data":{"provider":"p","tool":"lookup","tool_parameters":{"n":9007199254740993,"state":[false,0,{},[]]}}}`)
}

func TestDispatchPreservesAgentHistory(t *testing.T) {
	checkDispatchJSON[requests.RequestInvokeAgentStrategy](t, `{"tenant_id":"t","data":{"agent_strategy_provider":"p","agent_strategy":"fc","agent_strategy_params":{"model":{"history_prompt_messages":[{"role":"assistant","opaque_body":{"n":9007199254740993,"state":[false,0,{},[]]}}]}}}}`)
}
