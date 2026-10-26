package cache

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoinHostPort(t *testing.T) {
	assert.Equal(t, "127.0.0.1:6379", joinHostPort("127.0.0.1", "6379"))
	assert.Equal(t, "redis-master:6379", joinHostPort("redis-master", "6379"))
}

func TestEnsureRedisWritableRejectsReadOnlyReplica(t *testing.T) {
	err := InitRedisClient("127.0.0.1:6380", RedisCredentials{Password: "difyai123456"}, false, 0, nil)
	if err == nil {
		Close()
		t.Fatal("expected init to fail against a read-only replica")
	}
	if strings.Contains(err.Error(), "connection refused") {
		t.Skip("replica on :6380 not available (optional: integration/docker/docker-compose.sentinel.yml)")
	}
	// InitRedisClient should fail before returning when the node is read-only.
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "writable")
}
