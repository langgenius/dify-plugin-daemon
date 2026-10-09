//go:build integration

package cache

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func sentinelIntegrationAvailable(t *testing.T) {
	t.Helper()
	sentinel := redis.NewSentinelClient(&redis.Options{
		Addr:     "127.0.0.1:26379",
		Password: "difyai123456",
	})
	defer sentinel.Close()
	if _, err := sentinel.Ping(context.Background()).Result(); err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			t.Skip("sentinel stack not running (integration/docker/docker-compose.sentinel.yml)")
		}
		require.NoError(t, err)
	}
}

func TestInitRedisSentinelClientWritable(t *testing.T) {
	sentinelIntegrationAvailable(t)
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

func TestInitRedisSentinelClientDuringPromotionLag(t *testing.T) {
	sentinelIntegrationAvailable(t)
	t.Cleanup(func() { _ = Close() })

	origHook := discoverWritableMasterHook
	t.Cleanup(func() { discoverWritableMasterHook = origHook })

	var attempts int
	discoverWritableMasterHook = func(sentinels []string, o sentinelDiscoveryOptions) (string, map[string]string, error) {
		attempts++
		if attempts < 2 {
			return "", map[string]string{"127.0.0.1:6380": "slave"}, errNoWritableRedisMaster
		}
		discoverWritableMasterHook = nil
		return discoverWritableMaster(sentinels, o)
	}

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
	require.GreaterOrEqual(t, attempts, 2)
}

func TestRedisSentinelSurvivesFailover(t *testing.T) {
	if os.Getenv("RUN_SENTINEL_FAILOVER_TEST") != "1" {
		t.Skip("set RUN_SENTINEL_FAILOVER_TEST=1 to run destructive sentinel failover test")
	}
	sentinelIntegrationAvailable(t)
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

	sentinel := redis.NewSentinelClient(&redis.Options{
		Addr:     "127.0.0.1:26379",
		Password: "difyai123456",
	})
	defer sentinel.Close()

	require.NoError(t, sentinel.Do(context.Background(), "SENTINEL", "FAILOVER", "mymaster").Err())

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		_, err = SetNX("sentinel-failover-lock", "ok", redisMasterWriteProbeTTL)
		if err == nil {
			return
		}
		if !redis.IsReadOnlyError(err) {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("SetNX still failing after sentinel failover: %v", err)
}

func TestRedisSentinelIndependentSentinelCredentials(t *testing.T) {
	sentinelIntegrationAvailable(t)
	t.Cleanup(func() { _ = Close() })

	err := InitRedisSentinelClient(
		[]string{"127.0.0.1:26379"},
		"mymaster",
		RedisCredentials{Password: "difyai123456"},
		"",
		"wrong-sentinel-password",
		false,
		0,
		2,
		nil,
	)
	require.Error(t, err)
}
