package requests

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/langgenius/dify-plugin-daemon/pkg/utils/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFirstTokenBudget(t *testing.T) {
	assert.Equal(t, time.Duration(0), InvokeLLMSchema{Stream: true}.FirstTokenBudget())
	assert.Equal(t, time.Duration(0), InvokeLLMSchema{Stream: true, FirstTokenTimeout: -1}.FirstTokenBudget())
	assert.Equal(t, 1500*time.Millisecond, InvokeLLMSchema{Stream: true, FirstTokenTimeout: 1.5}.FirstTokenBudget())
}

func TestFirstTokenBudgetIsNotAppliedToANonStreamingCall(t *testing.T) {
	assert.Equal(t, time.Duration(0), InvokeLLMSchema{FirstTokenTimeout: 1.5}.FirstTokenBudget())
	assert.Equal(t, time.Duration(0), RequestStartPolling{
		RequestInvokeLLM: RequestInvokeLLM{
			InvokeLLMSchema: InvokeLLMSchema{FirstTokenTimeout: 1.5},
		},
	}.FirstTokenBudget())
}

func TestFirstTokenTimeoutIsDecodedFromTheWire(t *testing.T) {
	request := RequestInvokeLLM{}
	require.NoError(t, json.Unmarshal([]byte(`{"first_token_timeout": 2.5}`), &request))
	assert.InDelta(t, 2.5, request.FirstTokenTimeout, 1e-9)
}

func TestFirstTokenTimeoutIsHiddenFromPluginsThatWereNotGivenOne(t *testing.T) {
	payload := parser.StructToMap(RequestInvokeLLM{})

	_, present := payload["first_token_timeout"]
	assert.False(t, present)
}

func TestFirstTokenTimeoutIsForwardedToThePlugin(t *testing.T) {
	request := RequestInvokeLLM{}
	request.FirstTokenTimeout = 2.5

	payload := parser.StructToMap(request)

	assert.InDelta(t, 2.5, payload["first_token_timeout"], 1e-9)
}
