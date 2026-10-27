package local_runtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLookupSessionReturnsFalseForUnknownSession(t *testing.T) {
	r := &LocalPluginRuntime{}
	instance, ok := r.LookupSession("nonexistent-session-id")
	require.False(t, ok)
	require.Nil(t, instance)
}

func TestLookupSessionReturnsInstanceAfterStore(t *testing.T) {
	r := &LocalPluginRuntime{}
	stub := &PluginInstance{}
	r.sessionToInstanceMap.Store("session-A", stub)
	defer r.sessionToInstanceMap.Delete("session-A")

	instance, ok := r.LookupSession("session-A")
	require.True(t, ok)
	require.Same(t, stub, instance)
}

func TestLookupSessionReturnsFalseAfterDelete(t *testing.T) {
	r := &LocalPluginRuntime{}
	r.sessionToInstanceMap.Store("session-B", &PluginInstance{})
	r.sessionToInstanceMap.Delete("session-B")

	instance, ok := r.LookupSession("session-B")
	require.False(t, ok)
	require.Nil(t, instance)
}
