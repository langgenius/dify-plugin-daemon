package integration_test

import (
	_ "embed"
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
	testutils "github.com/langgenius/dify-plugin-daemon/internal/core/testutils"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/agent_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/routine"
	"github.com/stretchr/testify/assert"
)

//go:embed testdata/langgenius-agent_0.0.14.difypkg
var difyOfficialAgent []byte

//go:embed testdata/invoke_agent_strategy_json.json
var invokeAgentStrategyJson []byte

const (
	_testingPath = "./integration_test_cwd"
)

func TestDifyOfficialAgentIntegration(t *testing.T) {
	routine.InitPool(10000)

	defer testutils.ClearTestingPath(_testingPath)

	runtime, err := testutils.GetRuntime(difyOfficialAgent, _testingPath, 1)
	assert.NoError(t, err)

	invokePayload, err := parser.UnmarshalJsonBytes2Map(invokeAgentStrategyJson)
	assert.NoError(t, err)
	response, err := testutils.RunOnce[requests.RequestInvokeAgentStrategy, agent_entities.AgentStrategyResponseChunk](
		runtime,
		access_types.PLUGIN_ACCESS_TYPE_AGENT_STRATEGY,
		access_types.PLUGIN_ACCESS_ACTION_INVOKE_AGENT_STRATEGY,
		requests.RequestInvokeAgentStrategy{
			InvokeAgentStrategySchema: requests.InvokeAgentStrategySchema{
				AgentStrategyProvider: "agent",
				AgentStrategy:         "function_calling",
				AgentStrategyParams:   invokePayload,
			},
		},
	)

	assert.NoError(t, err)

	for response.Next() {
		_, err := response.Read()
		assert.NoError(t, err)
	}
}
