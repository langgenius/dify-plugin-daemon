package session_manager

import (
	"context"
	"testing"
	"unsafe"

	"github.com/langgenius/dify-plugin-daemon/internal/core/dify_invocation/mock"
	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
	"github.com/stretchr/testify/require"
)

var testPluginIdentifier = plugin_entities.PluginUniqueIdentifier("author/plugin:1.0.0@0123456789abcdef0123456789abcdef")

func TestGetSessionTraceInMemory(t *testing.T) {
	trace := log.TraceContext{TraceID: "trace-in-memory", SpanID: "span-in-memory"}
	identity := log.Identity{TenantID: "tenant-1", UserID: "user-1", UserType: "member"}
	ctx := log.WithIdentity(log.WithTrace(context.Background(), trace), identity)

	invocation := mock.NewMockedDifyInvocation()
	session := NewSession(NewSessionPayload{
		TenantID:               "tenant-1",
		UserID:                 "user-1",
		PluginUniqueIdentifier: testPluginIdentifier,
		ClusterID:              "cluster-1",
		InvokeFrom:             access_types.PLUGIN_ACCESS_TYPE_TOOL,
		Action:                 access_types.PLUGIN_ACCESS_ACTION_INVOKE_TOOL,
		BackwardsInvocation:    invocation,
		RequestContext:         ctx,
		IgnoreCache:            true, // non-serverless flow keeps everything in memory
	})
	t.Cleanup(func() {
		DeleteSession(DeleteSessionPayload{ID: session.ID, IgnoreCache: true})
	})

	got, err := GetSession(session.ID)
	require.NoError(t, err)
	require.Same(t, session, got)
	require.Equal(t, trace, got.TraceContext)
	require.Equal(t, identity, got.IdentityContext)

	// get address of the two contexts to ensure they are the same instance
	sessionCtxAddr := uintptr(unsafe.Pointer(&session.requestContext))
	gotCtxAddr := uintptr(unsafe.Pointer(&got.requestContext))
	require.Equal(t, sessionCtxAddr, gotCtxAddr) // should not be a copy

	traceFromSessionCtx, ok := log.TraceFromContext(got.RequestContext())
	require.True(t, ok)
	require.Equal(t, trace, traceFromSessionCtx)
	identityFromSessionCtx, ok := log.IdentityFromContext(got.RequestContext())
	require.True(t, ok)
	require.Equal(t, identity, identityFromSessionCtx)

	traceFromInvocation, ok := log.TraceFromContext(invocation.Context())
	require.True(t, ok)
	require.Equal(t, trace, traceFromInvocation)
	identityFromInvocation, ok := log.IdentityFromContext(invocation.Context())
	require.True(t, ok)
	require.Equal(t, identity, identityFromInvocation)
}

func TestGetSessionTraceFromDistributedCache(t *testing.T) {
	require.NoError(t, cache.InitRedisClient("127.0.0.1:6379", "", "difyai123456", false, 0))
	t.Cleanup(func() {
		cache.Close()
	})

	trace := log.TraceContext{TraceID: "trace-in-cache", SpanID: "span-in-cache"}
	identity := log.Identity{TenantID: "tenant-cache", UserID: "user-cache", UserType: "member"}
	ctx := log.WithIdentity(log.WithTrace(context.Background(), trace), identity)

	session := NewSession(NewSessionPayload{
		TenantID:               "tenant-cache",
		UserID:                 "user-cache",
		PluginUniqueIdentifier: testPluginIdentifier,
		ClusterID:              "cluster-cache",
		InvokeFrom:             access_types.PLUGIN_ACCESS_TYPE_TOOL,
		Action:                 access_types.PLUGIN_ACCESS_ACTION_INVOKE_TOOL,
		RequestContext:         ctx,
	})
	t.Cleanup(func() {
		DeleteSession(DeleteSessionPayload{ID: session.ID, IgnoreCache: false})
	})

	// simulate serverless node: remove local memory entry but keep redis cache
	DeleteSession(DeleteSessionPayload{ID: session.ID, IgnoreCache: true})

	cached, err := GetSession(session.ID)
	require.NoError(t, err)
	require.NotSame(t, session, cached)

	require.Equal(t, trace, cached.TraceContext)
	require.Equal(t, identity, cached.IdentityContext)

	traceFromCtx, ok := log.TraceFromContext(cached.RequestContext())
	require.True(t, ok)
	require.Equal(t, trace, traceFromCtx)
	identityFromCtx, ok := log.IdentityFromContext(cached.RequestContext())
	require.True(t, ok)
	require.Equal(t, identity, identityFromCtx)

	// get address of the two contexts to ensure they are different copies
	sessionCtxAddr := uintptr(unsafe.Pointer(&session.requestContext))
	cachedCtxAddr := uintptr(unsafe.Pointer(&cached.requestContext))

	require.NotEqual(t, sessionCtxAddr, cachedCtxAddr) // should be a copy

	// attaching a backwards invocation after cache load should still carry the trace context
	invocation := mock.NewMockedDifyInvocation()
	cached.BindBackwardsInvocation(invocation)

	traceFromInvocation, ok := log.TraceFromContext(invocation.Context())
	require.True(t, ok)
	require.Equal(t, trace, traceFromInvocation)
	identityFromInvocation, ok := log.IdentityFromContext(invocation.Context())
	require.True(t, ok)
	require.Equal(t, identity, identityFromInvocation)
}
