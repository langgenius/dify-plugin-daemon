package cache

import (
	"errors"
	"sync/atomic"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errReadonlyReplica = errors.New("READONLY You can't write against a read only replica.")

func TestSortCandidatesBySentinelVotes(t *testing.T) {
	candidates := map[string]struct{}{
		"10.0.0.1:6379": {},
		"10.0.0.2:6379": {},
		"10.0.0.3:6379": {},
	}
	votes := map[string]int{
		"10.0.0.1:6379": 1,
		"10.0.0.2:6379": 3,
		"10.0.0.3:6379": 2,
	}
	sorted := sortCandidatesBySentinelVotes(candidates, votes)
	require.Equal(t, []string{"10.0.0.2:6379", "10.0.0.3:6379", "10.0.0.1:6379"}, sorted)
}

func TestTryWritableMasterPrefersSentinelQuorumOrder(t *testing.T) {
	origRole := nodeRoleFn
	origWrite := nodeWriteFn
	t.Cleanup(func() {
		nodeRoleFn = origRole
		nodeWriteFn = origWrite
	})

	nodeRoleFn = func(_ string, _ *redis.Options) (string, error) {
		return "master", nil
	}
	var writeCalls atomic.Int32
	nodeWriteFn = func(addr string, _ *redis.Options) error {
		if addr == "10.0.0.2:6379" && writeCalls.Add(1) == 1 {
			return errReadonlyReplica
		}
		return nil
	}

	o := sentinelDiscoveryOptions{}
	addrs := []string{"10.0.0.2:6379", "10.0.0.1:6379"}
	master, roles, err := tryWritableMasterFromCandidates(addrs, o)
	require.NoError(t, err)
	assert.Equal(t, "10.0.0.1:6379", master)
	assert.Contains(t, roles["10.0.0.2:6379"], "write failed")
}

func TestDiscoverWritableMasterWithRetry(t *testing.T) {
	origHook := discoverWritableMasterHook
	t.Cleanup(func() {
		discoverWritableMasterHook = origHook
	})

	var attempts atomic.Int32
	discoverWritableMasterHook = func(_ []string, _ sentinelDiscoveryOptions) (string, map[string]string, error) {
		if attempts.Add(1) < 3 {
			return "", map[string]string{"10.0.0.1:6379": "slave"}, errNoWritableRedisMaster
		}
		return "10.0.0.2:6379", map[string]string{"10.0.0.2:6379": "master"}, nil
	}

	master, _, err := discoverWritableMasterWithRetry(nil, sentinelDiscoveryOptions{})
	require.NoError(t, err)
	assert.Equal(t, "10.0.0.2:6379", master)
	assert.Equal(t, int32(3), attempts.Load())
}

func TestDiscoverWritableMasterWithRetryExhaustsDeadline(t *testing.T) {
	origHook := discoverWritableMasterHook
	t.Cleanup(func() {
		discoverWritableMasterHook = origHook
	})

	discoverWritableMasterHook = func(_ []string, _ sentinelDiscoveryOptions) (string, map[string]string, error) {
		return "", map[string]string{"10.0.0.1:6379": "slave"}, errNoWritableRedisMaster
	}

	_, roles, err := discoverWritableMasterWithRetry(nil, sentinelDiscoveryOptions{})
	require.Error(t, err)
	assert.ErrorIs(t, err, errNoWritableRedisMaster)
	assert.Contains(t, err.Error(), "slave")
	assert.Contains(t, roles, "10.0.0.1:6379")
}

func TestInconsistentSentinelQuorumPicksMajorityWritable(t *testing.T) {
	origRole := nodeRoleFn
	origWrite := nodeWriteFn
	t.Cleanup(func() {
		nodeRoleFn = origRole
		nodeWriteFn = origWrite
	})

	nodeRoleFn = func(_ string, _ *redis.Options) (string, error) {
		return "master", nil
	}
	nodeWriteFn = func(_ string, _ *redis.Options) error {
		return nil
	}

	candidates := map[string]struct{}{
		"old-master:6379": {},
		"new-master:6379": {},
	}
	votes := map[string]int{
		"old-master:6379": 1,
		"new-master:6379": 2,
	}
	sorted := sortCandidatesBySentinelVotes(candidates, votes)
	master, _, err := tryWritableMasterFromCandidates(sorted, sentinelDiscoveryOptions{})
	require.NoError(t, err)
	assert.Equal(t, "new-master:6379", master)
}

func TestRediscoverRequiresRuntimeConfig(t *testing.T) {
	sentinelRuntimeMu.Lock()
	prev := sentinelRuntime
	sentinelRuntime = nil
	sentinelRuntimeMu.Unlock()
	t.Cleanup(func() {
		sentinelRuntimeMu.Lock()
		sentinelRuntime = prev
		sentinelRuntimeMu.Unlock()
	})

	err := rediscoverSentinelClient()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
}
