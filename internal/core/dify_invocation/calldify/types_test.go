package calldify

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithContextReturnsIndependentInvocations(t *testing.T) {
	base := &RealBackwardsInvocation{}
	canceledContext, cancel := context.WithCancel(context.Background())
	first := base.WithContext(canceledContext)
	second := base.WithContext(context.Background())
	cancel()

	require.ErrorIs(t, first.Context().Err(), context.Canceled)
	require.NoError(t, second.Context().Err())
	require.NoError(t, base.Context().Err())
}
