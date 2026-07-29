package definitions

import (
	"testing"

	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
)

func TestEmbeddingDispatchersUseServerlessAwareHandlers(t *testing.T) {
	want := map[string]string{
		"InvokeTextEmbedding":       "invokeTextEmbeddingServerlessAware",
		"GetTextEmbeddingNumTokens": "getTextEmbeddingNumTokensServerlessAware",
	}
	found := map[string]string{}

	for _, dispatcher := range PluginDispatchers {
		if dispatcher.CustomTunnelHandler != "" {
			found[dispatcher.Name] = dispatcher.CustomTunnelHandler
		}
	}

	if len(found) != len(want) {
		t.Fatalf("custom handler count = %d, want %d: %#v", len(found), len(want), found)
	}
	for name, handler := range want {
		if found[name] != handler {
			t.Errorf("dispatcher %s custom handler = %q, want %q", name, found[name], handler)
		}
	}
}

func TestMultimodalDispatchersPreserveDedicatedActions(t *testing.T) {
	want := map[string]access_types.PluginAccessAction{
		"InvokeMultimodalEmbedding": access_types.PLUGIN_ACCESS_ACTION_INVOKE_MULTIMODAL_EMBEDDING,
		"InvokeMultimodalRerank":    access_types.PLUGIN_ACCESS_ACTION_INVOKE_MULTIMODAL_RERANK,
	}
	found := map[string]access_types.PluginAccessAction{}

	for _, dispatcher := range PluginDispatchers {
		if _, ok := want[dispatcher.Name]; ok {
			found[dispatcher.Name] = dispatcher.AccessAction
		}
	}

	for name, action := range want {
		if found[name] != action {
			t.Errorf("dispatcher %s action = %q, want %q", name, found[name], action)
		}
	}
}
