package cache

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/redis/go-redis/v9"
)

var (
	client *redis.Client
	ctx    = context.Background()

	ErrDBNotInit = errors.New("redis client not init")
)

func InitRedisClient(addr, password string) error {
	client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
	})

	if _, err := client.Ping(ctx).Result(); err != nil {
		return err
	}

	return nil
}

func serialKey(keys ...string) string {
	return strings.Join(append(
		[]string{"plugin_daemon"},
		keys...,
	), ":")
}

func Store(key string, value any, time time.Duration) error {
	if client == nil {
		return ErrDBNotInit
	}

	return client.Set(ctx, serialKey(key), value, time).Err()
}

func Get[T any](key string) (*T, error) {
	if client == nil {
		return nil, ErrDBNotInit
	}

	val, err := client.Get(ctx, serialKey(key)).Result()
	if err != nil {
		return nil, err
	}

	result, err := parser.UnmarshalJson[T](val)
	return &result, err
}

func Del(key string) error {
	if client == nil {
		return ErrDBNotInit
	}

	return client.Del(ctx, serialKey(key)).Err()
}

func Exist(key string) (int64, error) {
	if client == nil {
		return 0, ErrDBNotInit
	}

	return client.Exists(ctx, serialKey(key)).Result()
}

func Increase(key string) (int64, error) {
	if client == nil {
		return 0, ErrDBNotInit
	}

	return client.Incr(ctx, serialKey(key)).Result()
}

func Decrease(key string) (int64, error) {
	if client == nil {
		return 0, ErrDBNotInit
	}

	return client.Decr(ctx, serialKey(key)).Result()
}

func SetExpire(key string, time time.Duration) error {
	if client == nil {
		return ErrDBNotInit
	}

	return client.Expire(ctx, serialKey(key), time).Err()
}

func SetMapField(key string, v map[string]any) error {
	if client == nil {
		return ErrDBNotInit
	}

	return client.HMSet(ctx, serialKey(key), v).Err()
}

func SetMapOneField(key string, field string, value any) error {
	if client == nil {
		return ErrDBNotInit
	}

	return client.HSet(ctx, serialKey(key), field, value).Err()
}

func GetMapField[T any](key string, field string) (*T, error) {
	if client == nil {
		return nil, ErrDBNotInit
	}

	val, err := client.HGet(ctx, serialKey(key), field).Result()
	if err != nil {
		return nil, err
	}

	result, err := parser.UnmarshalJson[T](val)
	return &result, err
}

func DelMapField(key string, field string) error {
	if client == nil {
		return ErrDBNotInit
	}

	return client.HDel(ctx, serialKey(key), field).Err()
}

func GetMap[V any](key string) (map[string]V, error) {
	if client == nil {
		return nil, ErrDBNotInit
	}

	val, err := client.HGetAll(ctx, serialKey(key)).Result()
	if err != nil {
		return nil, err
	}

	result := make(map[string]V)
	for k, v := range val {
		value, err := parser.UnmarshalJson[V](v)
		if err != nil {
			return nil, err
		}

		result[k] = value
	}

	return result, nil
}

var (
	ErrLockTimeout = errors.New("lock timeout")
)

// Lock key, expire time takes responsibility for expiration time
// try_lock_timeout takes responsibility for the timeout of trying to lock
func Lock(key string, expire time.Duration, try_lock_timeout time.Duration) error {
	if client == nil {
		return ErrDBNotInit
	}

	const LOCK_DURATION = 20 * time.Millisecond

	ticker := time.NewTicker(LOCK_DURATION)
	defer ticker.Stop()

	for range ticker.C {
		if _, err := client.SetNX(ctx, serialKey(key), "1", expire).Result(); err == nil {
			return nil
		}

		try_lock_timeout -= LOCK_DURATION
		if try_lock_timeout <= 0 {
			return ErrLockTimeout
		}
	}

	return nil
}

func Unlock(key string) error {
	if client == nil {
		return ErrDBNotInit
	}

	return client.Del(ctx, serialKey(key)).Err()
}
