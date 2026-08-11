package cache

import (
	"errors"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"
)

func setupOwnedLockTest(t *testing.T) *miniredis.Miniredis {
	t.Helper()

	server := miniredis.RunT(t)
	require.NoError(t, InitRedisClient(server.Addr(), RedisCredentials{}, false, 0, nil))
	t.Cleanup(func() {
		_ = Close()
	})
	return server
}

func TestOwnedLockRejectsStaleOwnerUnlock(t *testing.T) {
	server := setupOwnedLockTest(t)

	first, err := AcquireOwnedLock("owned-lock", time.Second, time.Second)
	require.NoError(t, err)

	server.FastForward(time.Second + time.Millisecond)

	second, err := AcquireOwnedLock("owned-lock", time.Second, time.Second)
	require.NoError(t, err)
	require.ErrorIs(t, first.Unlock(), ErrLockNotOwned)

	value, err := server.Get(serialKey("owned-lock"))
	require.NoError(t, err)
	require.Equal(t, second.token, value)
	require.NoError(t, second.Unlock())
}

func TestOwnedLockOnlyRenewsCurrentOwner(t *testing.T) {
	server := setupOwnedLockTest(t)

	first, err := AcquireOwnedLock("renewable-lock", time.Second, time.Second)
	require.NoError(t, err)
	server.FastForward(time.Second + time.Millisecond)

	second, err := AcquireOwnedLock("renewable-lock", time.Second, time.Second)
	require.NoError(t, err)
	require.ErrorIs(t, first.Renew(2*time.Second), ErrLockNotOwned)
	require.NoError(t, second.Renew(2*time.Second))

	server.FastForward(time.Second + time.Millisecond)
	value, err := server.Get(serialKey("renewable-lock"))
	require.NoError(t, err)
	require.Equal(t, second.token, value)
	require.NoError(t, second.Unlock())
}

func TestOwnedLockValidatesExpiration(t *testing.T) {
	setupOwnedLockTest(t)

	_, err := AcquireOwnedLock("invalid-lock", 0, time.Second)
	require.ErrorIs(t, err, ErrInvalidLockExpiration)

	lock, err := AcquireOwnedLock("valid-lock", time.Second, time.Second)
	require.NoError(t, err)
	require.True(t, errors.Is(lock.Renew(0), ErrInvalidLockExpiration))
	require.NoError(t, lock.Unlock())
}
