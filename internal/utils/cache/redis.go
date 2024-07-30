package cache

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/redis/go-redis/v9"
)

var (
	client *redis.Client
	ctx    = context.Background()

	ErrDBNotInit = errors.New("redis client not init")
	ErrNotFound  = errors.New("key not found")
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

func Close() error {
	if client == nil {
		return ErrDBNotInit
	}

	return client.Close()
}

func getCmdable(context ...redis.Cmdable) redis.Cmdable {
	if len(context) > 0 {
		return context[0]
	}

	return client
}

func serialKey(keys ...string) string {
	return strings.Join(append(
		[]string{"plugin_daemon"},
		keys...,
	), ":")
}

func Store(key string, value any, time time.Duration, context ...redis.Cmdable) error {
	if client == nil {
		return ErrDBNotInit
	}

	if _, ok := value.(string); !ok {
		value = parser.MarshalJson(value)
	}

	return getCmdable(context...).Set(ctx, serialKey(key), value, time).Err()
}

func Get[T any](key string, context ...redis.Cmdable) (*T, error) {
	if client == nil {
		return nil, ErrDBNotInit
	}

	val, err := getCmdable(context...).Get(ctx, serialKey(key)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if val == "" {
		return nil, ErrNotFound
	}

	result, err := parser.UnmarshalJson[T](val)
	return &result, err
}

func GetString(key string, context ...redis.Cmdable) (string, error) {
	if client == nil {
		return "", ErrDBNotInit
	}

	v, err := getCmdable(context...).Get(ctx, serialKey(key)).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ErrNotFound
		}
	}

	return v, err
}

func Del(key string, context ...redis.Cmdable) error {
	if client == nil {
		return ErrDBNotInit
	}

	_, err := getCmdable(context...).Del(ctx, serialKey(key)).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrNotFound
		}

		return err
	}

	return nil
}

func Exist(key string, context ...redis.Cmdable) (int64, error) {
	if client == nil {
		return 0, ErrDBNotInit
	}

	return getCmdable(context...).Exists(ctx, serialKey(key)).Result()
}

func Increase(key string, context ...redis.Cmdable) (int64, error) {
	if client == nil {
		return 0, ErrDBNotInit
	}

	num, err := getCmdable(context...).Incr(ctx, serialKey(key)).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, ErrNotFound
		}
		return 0, err
	}

	return num, nil
}

func Decrease(key string, context ...redis.Cmdable) (int64, error) {
	if client == nil {
		return 0, ErrDBNotInit
	}

	return getCmdable(context...).Decr(ctx, serialKey(key)).Result()
}

func SetExpire(key string, time time.Duration, context ...redis.Cmdable) error {
	if client == nil {
		return ErrDBNotInit
	}

	return getCmdable(context...).Expire(ctx, serialKey(key), time).Err()
}

func SetMapField(key string, v map[string]any, context ...redis.Cmdable) error {
	if client == nil {
		return ErrDBNotInit
	}

	return getCmdable(context...).HMSet(ctx, serialKey(key), v).Err()
}

func SetMapOneField(key string, field string, value any, context ...redis.Cmdable) error {
	if client == nil {
		return ErrDBNotInit
	}

	if _, ok := value.(string); !ok {
		value = parser.MarshalJson(value)
	}

	return getCmdable(context...).HSet(ctx, serialKey(key), field, value).Err()
}

func GetMapField[T any](key string, field string, context ...redis.Cmdable) (*T, error) {
	if client == nil {
		return nil, ErrDBNotInit
	}

	val, err := getCmdable(context...).HGet(ctx, serialKey(key), field).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrNotFound
		}
		return nil, err
	}

	result, err := parser.UnmarshalJson[T](val)
	return &result, err
}

func DelMapField(key string, field string, context ...redis.Cmdable) error {
	if client == nil {
		return ErrDBNotInit
	}

	return getCmdable(context...).HDel(ctx, serialKey(key), field).Err()
}

func GetMap[V any](key string, context ...redis.Cmdable) (map[string]V, error) {
	if client == nil {
		return nil, ErrDBNotInit
	}

	val, err := getCmdable(context...).HGetAll(ctx, serialKey(key)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrNotFound
		}
		return nil, err
	}

	result := make(map[string]V)
	for k, v := range val {
		value, err := parser.UnmarshalJson[V](v)
		if err != nil {
			continue
		}

		result[k] = value
	}

	return result, nil
}

func ScanMap[V any](key string, prefix string, context ...redis.Cmdable) (map[string]V, error) {
	if client == nil {
		return nil, ErrDBNotInit
	}

	result := make(map[string]V)

	ScanMapAsync[V](key, prefix, func(m map[string]V) error {
		for k, v := range m {
			result[k] = v
		}

		return nil
	})

	return result, nil
}

func ScanMapAsync[V any](key string, prefix string, fn func(map[string]V) error, context ...redis.Cmdable) error {
	if client == nil {
		return ErrDBNotInit
	}

	cursor := uint64(0)

	for {
		kvs, new_cursor, err := getCmdable(context...).
			HScan(ctx, serialKey(key), cursor, fmt.Sprintf("%s*", prefix), 32).
			Result()

		if err != nil {
			return err
		}

		result := make(map[string]V)
		for i := 0; i < len(kvs); i += 2 {
			value, err := parser.UnmarshalJson[V](kvs[i+1])
			if err != nil {
				continue
			}

			result[kvs[i]] = value
		}

		if err := fn(result); err != nil {
			return err
		}

		if new_cursor == 0 {
			break
		}

		cursor = new_cursor
	}

	return nil
}

func SetNX[T any](key string, value T, expire time.Duration, context ...redis.Cmdable) (bool, error) {
	if client == nil {
		return false, ErrDBNotInit
	}

	return getCmdable(context...).SetNX(ctx, serialKey(key), value, expire).Result()
}

var (
	ErrLockTimeout = errors.New("lock timeout")
)

// Lock key, expire time takes responsibility for expiration time
// try_lock_timeout takes responsibility for the timeout of trying to lock
func Lock(key string, expire time.Duration, try_lock_timeout time.Duration, context ...redis.Cmdable) error {
	if client == nil {
		return ErrDBNotInit
	}

	const LOCK_DURATION = 20 * time.Millisecond

	ticker := time.NewTicker(LOCK_DURATION)
	defer ticker.Stop()

	for range ticker.C {
		if _, err := getCmdable(context...).SetNX(ctx, serialKey(key), "1", expire).Result(); err == nil {
			return nil
		}

		try_lock_timeout -= LOCK_DURATION
		if try_lock_timeout <= 0 {
			return ErrLockTimeout
		}
	}

	return nil
}

func Unlock(key string, context ...redis.Cmdable) error {
	if client == nil {
		return ErrDBNotInit
	}

	return getCmdable(context...).Del(ctx, serialKey(key)).Err()
}

func Expire(key string, time time.Duration, context ...redis.Cmdable) (bool, error) {
	if client == nil {
		return false, ErrDBNotInit
	}

	return getCmdable(context...).Expire(ctx, serialKey(key), time).Result()
}

func Transaction(fn func(redis.Pipeliner) error) error {
	if client == nil {
		return ErrDBNotInit
	}

	return client.Watch(ctx, func(tx *redis.Tx) error {
		_, err := tx.TxPipelined(ctx, func(p redis.Pipeliner) error {
			return fn(p)
		})
		if err == redis.Nil {
			return nil
		}
		return err
	})
}
