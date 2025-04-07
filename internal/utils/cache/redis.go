package cache

import (
	"context"
	"crypto/tls"
	"errors"
	"strings"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/redis/go-redis/v9"
)

var (
	client *redis.Client
	ctx    = context.Background()

	ErrDBNotInit = errors.New("redis client not init")
	ErrNotFound  = errors.New("key not found")
)

func getRedisOptions(addr, password string, useSsl bool, db int) *redis.Options {
	opts := &redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	}
	if useSsl {
		opts.TLSConfig = &tls.Config{}
	}
	return opts
}

func InitRedisClient(addr, password string, useSsl bool, db int) error {
	opts := getRedisOptions(addr, password, useSsl, db)
	client = redis.NewClient(opts)

	if _, err := client.Ping(ctx).Result(); err != nil {
		return err
	}

	return nil
}

// Close the redis client
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

// Store the key-value pair
func Store(key string, value any, time time.Duration, context ...redis.Cmdable) error {
	return store(serialKey(key), value, time, context...)
}

// store the key-value pair, without serialKey
func store(key string, value any, time time.Duration, context ...redis.Cmdable) error {
	if client == nil {
		return ErrDBNotInit
	}

	if _, ok := value.(string); !ok {
		var err error
		value, err = parser.MarshalCBOR(value)
		if err != nil {
			return err
		}
	}

	return getCmdable(context...).Set(ctx, key, value, time).Err()
}

// Get the value with key
func Get[T any](key string, context ...redis.Cmdable) (*T, error) {
	return get[T](serialKey(key), context...)
}

func get[T any](key string, context ...redis.Cmdable) (*T, error) {
	if client == nil {
		return nil, ErrDBNotInit
	}

	val, err := getCmdable(context...).Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if len(val) == 0 {
		return nil, ErrNotFound
	}

	result, err := parser.UnmarshalCBOR[T](val)
	return &result, err
}

// GetString get the string with key
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

// Del the key
func Del(key string, context ...redis.Cmdable) error {
	return del(serialKey(key), context...)
}

func del(key string, context ...redis.Cmdable) error {
	if client == nil {
		return ErrDBNotInit
	}

	_, err := getCmdable(context...).Del(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrNotFound
		}

		return err
	}

	return nil
}

// Exist check the key exist or not
func Exist(key string, context ...redis.Cmdable) (int64, error) {
	if client == nil {
		return 0, ErrDBNotInit
	}

	return getCmdable(context...).Exists(ctx, serialKey(key)).Result()
}

// Increase the key
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

// Decrease the key
func Decrease(key string, context ...redis.Cmdable) (int64, error) {
	if client == nil {
		return 0, ErrDBNotInit
	}

	return getCmdable(context...).Decr(ctx, serialKey(key)).Result()
}

// SetExpire set the expire time for the key
func SetExpire(key string, time time.Duration, context ...redis.Cmdable) error {
	if client == nil {
		return ErrDBNotInit
	}

	return getCmdable(context...).Expire(ctx, serialKey(key), time).Err()
}

// SetMapField set the map field with key
func SetMapField(key string, v map[string]any, context ...redis.Cmdable) error {
	if client == nil {
		return ErrDBNotInit
	}

	return getCmdable(context...).HMSet(ctx, serialKey(key), v).Err()
}

// SetMapOneField set the map field with key
func SetMapOneField(key string, field string, value any, context ...redis.Cmdable) error {
	if client == nil {
		return ErrDBNotInit
	}

	if _, ok := value.(string); !ok {
		value = parser.MarshalJson(value)
	}

	return getCmdable(context...).HSet(ctx, serialKey(key), field, value).Err()
}

// GetMapField get the map field with key
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

// GetMapFieldString get the string
func GetMapFieldString(key string, field string, context ...redis.Cmdable) (string, error) {
	if client == nil {
		return "", ErrDBNotInit
	}

	val, err := getCmdable(context...).HGet(ctx, serialKey(key), field).Result()
	if err != nil {
		if err == redis.Nil {
			return "", ErrNotFound
		}
		return "", err
	}

	return val, nil
}

// DelMapField delete the map field with key
func DelMapField(key string, field string, context ...redis.Cmdable) error {
	if client == nil {
		return ErrDBNotInit
	}

	return getCmdable(context...).HDel(ctx, serialKey(key), field).Err()
}

// GetMap get the map with key
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

// ScanKeys scan the keys with match pattern
func ScanKeys(match string, context ...redis.Cmdable) ([]string, error) {
	if client == nil {
		return nil, ErrDBNotInit
	}

	result := make([]string, 0)

	if err := ScanKeysAsync(match, func(keys []string) error {
		result = append(result, keys...)
		return nil
	}); err != nil {
		return nil, err
	}

	return result, nil
}

// ScanKeysAsync scan the keys with match pattern, format like "key*"
func ScanKeysAsync(match string, fn func([]string) error, context ...redis.Cmdable) error {
	if client == nil {
		return ErrDBNotInit
	}

	cursor := uint64(0)

	for {
		keys, newCursor, err := getCmdable(context...).Scan(ctx, cursor, match, 32).Result()
		if err != nil {
			return err
		}

		if err := fn(keys); err != nil {
			return err
		}

		if newCursor == 0 {
			break
		}

		cursor = newCursor
	}

	return nil
}

// ScanMap scan the map with match pattern, format like "key*"
func ScanMap[V any](key string, match string, context ...redis.Cmdable) (map[string]V, error) {
	if client == nil {
		return nil, ErrDBNotInit
	}

	result := make(map[string]V)

	ScanMapAsync[V](key, match, func(m map[string]V) error {
		for k, v := range m {
			result[k] = v
		}

		return nil
	})

	return result, nil
}

// ScanMapAsync scan the map with match pattern, format like "key*"
func ScanMapAsync[V any](key string, match string, fn func(map[string]V) error, context ...redis.Cmdable) error {
	if client == nil {
		return ErrDBNotInit
	}

	cursor := uint64(0)

	for {
		kvs, newCursor, err := getCmdable(context...).
			HScan(ctx, serialKey(key), cursor, match, 32).
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

		if newCursor == 0 {
			break
		}

		cursor = newCursor
	}

	return nil
}

// SetNX set the key-value pair with expire time
func SetNX[T any](key string, value T, expire time.Duration, context ...redis.Cmdable) (bool, error) {
	if client == nil {
		return false, ErrDBNotInit
	}

	// marshal the value
	bytes, err := parser.MarshalCBOR(value)
	if err != nil {
		return false, err
	}

	return getCmdable(context...).SetNX(ctx, serialKey(key), bytes, expire).Result()
}

var (
	ErrLockTimeout = errors.New("lock timeout")
)

// Lock key, expire time takes responsibility for expiration time
// try_lock_timeout takes responsibility for the timeout of trying to lock
func Lock(key string, expire time.Duration, tryLockTimeout time.Duration, context ...redis.Cmdable) error {
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

		tryLockTimeout -= LOCK_DURATION
		if tryLockTimeout <= 0 {
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

func Publish(channel string, message any, context ...redis.Cmdable) error {
	if client == nil {
		return ErrDBNotInit
	}

	if _, ok := message.(string); !ok {
		message = parser.MarshalJson(message)
	}

	return getCmdable(context...).Publish(ctx, channel, message).Err()
}

func Subscribe[T any](channel string) (<-chan T, func()) {
	pubsub := client.Subscribe(ctx, channel)
	ch := make(chan T)
	connectionEstablished := make(chan bool)

	go func() {
		defer close(ch)
		defer close(connectionEstablished)

		alive := true
		for alive {
			iface, err := pubsub.Receive(context.Background())
			if err != nil {
				log.Error("failed to receive message from redis: %s, will retry in 1 second", err.Error())
				time.Sleep(1 * time.Second)
				continue
			}
			switch data := iface.(type) {
			case *redis.Subscription:
				connectionEstablished <- true
			case *redis.Message:
				v, err := parser.UnmarshalJson[T](data.Payload)
				if err != nil {
					continue
				}

				ch <- v
			case *redis.Pong:
			default:
				alive = false
			}
		}
	}()

	// wait for the connection to be established
	<-connectionEstablished

	return ch, func() {
		pubsub.Close()
	}
}
