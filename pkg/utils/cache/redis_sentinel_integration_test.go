//go:build integration

package cache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitRedisSentinelClientWritable(t *testing.T) {
	t.Cleanup(func() { _ = Close() })

	err := InitRedisSentinelClient(
		[]string{"127.0.0.1:26379"},
		"mymaster",
		RedisCredentials{Password: "difyai123456"},
		"",
		"difyai123456",
		false,
		0,
		2,
		nil,
	)
	require.NoError(t, err)

	_, err = SetNX("sentinel-integration-lock", "ok", redisMasterWriteProbeTTL)
	require.NoError(t, err)
}
