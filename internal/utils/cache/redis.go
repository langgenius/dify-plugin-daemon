package cache

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/parser"
	"github.com/redis/go-redis/v9"
)

var (
	redisClient redis.Cmdable
	ctx         = context.Background()

	ErrDBNotInit = errors.New("redis client not init")
	ErrNotFound  = errors.New("key not found")
)

// RedisInterface is an abstract interface for Redis client
type RedisInterface interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Exists(ctx context.Context, keys ...string) *redis.IntCmd
	Incr(ctx context.Context, key string) *redis.IntCmd
	Decr(ctx context.Context, key string) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	HMSet(ctx context.Context, key string, values ...interface{}) *redis.StatusCmd
	HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
	HGet(ctx context.Context, key, field string) *redis.StringCmd
	HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd
	HGetAll(ctx context.Context, key string) *redis.MapStringStringCmd
	Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd
	HScan(ctx context.Context, key string, cursor uint64, match string, count int64) *redis.ScanCmd
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd
	Subscribe(ctx context.Context, channels ...string) *redis.PubSub
	Pipeline() redis.Pipeliner
	Close() error
}

// StandaloneRedis wraps the standalone Redis client
type StandaloneRedis struct {
	*redis.Client
}

// ClusterRedis wraps the cluster Redis client
type ClusterRedis struct {
	*redis.ClusterClient
}

// Pipeline wraps the Pipeline method to ensure ClusterRedis implements the interface
func (c *ClusterRedis) Pipeline() redis.Pipeliner {
	return c.ClusterClient.Pipeline()
}

// Create Redis standalone config options
func getRedisOptions(addr, username, password string, useSsl bool, db int) *redis.Options {
	opts := &redis.Options{
		Addr:     addr,
		Username: username,
		Password: password,
		DB:       db,
	}
	if useSsl {
		opts.TLSConfig = &tls.Config{}
	}
	return opts
}

// InitStandaloneClient initializes a standalone Redis client (internal use)
func InitStandaloneClient(addr, username, password string, useSsl bool, db int) (redis.Cmdable, error) {
	opts := getRedisOptions(addr, username, password, useSsl, db)
	client := redis.NewClient(opts)

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	return client, nil
}

// InitClusterClient initializes a cluster Redis client (internal use)
func InitClusterClient(addrs []string, password string, useSsl bool) (redis.Cmdable, error) {
	opts := &redis.ClusterOptions{
		Addrs:    addrs,
		Password: password,
	}

	if useSsl {
		opts.TLSConfig = &tls.Config{}
	}

	client := redis.NewClusterClient(opts)

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	return client, nil
}

// InitSentinelClient initializes a sentinel mode Redis client (internal use)
func InitSentinelClient(
	masterName string,
	sentinelAddrs []string,
	username string, // Username for Redis master-slave servers
	password string, // Password for Redis master-slave servers
	sentinelUsername string, // Username for connecting to sentinel servers
	sentinelPassword string, // Password for connecting to sentinel servers
	useSsl bool,
	db int,
	socketTimeout float64,
) (redis.Cmdable, error) {
	opts := &redis.FailoverOptions{
		MasterName:       masterName,
		SentinelAddrs:    sentinelAddrs,
		Username:         username,         // Username for Redis master-slave servers
		Password:         password,         // Password for Redis master-slave servers
		SentinelUsername: sentinelUsername, // Username for connecting to sentinel servers
		SentinelPassword: sentinelPassword, // Password for connecting to sentinel servers
		DB:               db,
	}

	// Set socket timeout
	if socketTimeout > 0 {
		opts.DialTimeout = time.Duration(socketTimeout * float64(time.Second))
	}

	if useSsl {
		opts.TLSConfig = &tls.Config{}
	}

	client := redis.NewFailoverClient(opts)

	// Print client type for debugging
	log.Info("Redis failover client type: %T", client)

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	return client, nil
}

// InitRedisClient initializes a standalone Redis client
func InitRedisClient(addr, username, password string, useSsl bool, db int) error {
	client, err := InitStandaloneClient(addr, username, password, useSsl, db)
	if err != nil {
		return err
	}

	redisClient = client
	return nil
}

// InitRedisClusterClient initializes a cluster Redis client
func InitRedisClusterClient(addrs []string, password string, useSsl bool) error {
	client, err := InitClusterClient(addrs, password, useSsl)
	if err != nil {
		return err
	}

	redisClient = client
	return nil
}

// InitRedisSentinelClient initializes a sentinel mode Redis client
func InitRedisSentinelClient(
	masterName string,
	sentinelAddrs []string,
	username string,
	password string,
	sentinelUsername string,
	sentinelPassword string,
	useSsl bool,
	db int,
	socketTimeout float64,
) error {
	client, err := InitSentinelClient(
		masterName,
		sentinelAddrs,
		username,
		password,
		sentinelUsername,
		sentinelPassword,
		useSsl,
		db,
		socketTimeout,
	)
	if err != nil {
		return err
	}

	redisClient = client
	return nil
}

// Close closes the Redis client
func Close() error {
	if redisClient == nil {
		return ErrDBNotInit
	}

	// Close according to different client types
	switch client := redisClient.(type) {
	case *redis.Client:
		return client.Close()
	case *redis.ClusterClient:
		return client.Close()
	default:
		return fmt.Errorf("unknown redis client type: %T", client)
	}
}

func getCmdable(context ...redis.Cmdable) redis.Cmdable {
	if len(context) > 0 {
		return context[0]
	}

	if redisClient == nil {
		return nil
	}

	return redisClient
}

func serialKey(keys ...string) string {
	return strings.Join(append(
		[]string{"plugin_daemon"},
		keys...,
	), ":")
}

// Store stores key-value pair
func Store(key string, value any, time time.Duration, context ...redis.Cmdable) error {
	return store(serialKey(key), value, time, context...)
}

// store stores key-value pair without serializing key
func store(key string, value any, time time.Duration, context ...redis.Cmdable) error {
	if redisClient == nil {
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

// Get retrieves a value by key
func Get[T any](key string, context ...redis.Cmdable) (*T, error) {
	return get[T](serialKey(key), context...)
}

func get[T any](key string, context ...redis.Cmdable) (*T, error) {
	if redisClient == nil {
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

// GetString retrieves a string value by key
func GetString(key string, context ...redis.Cmdable) (string, error) {
	if redisClient == nil {
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

// Del deletes a key
func Del(key string, context ...redis.Cmdable) (int64, error) {
	return del(serialKey(key), context...)
}

func del(key string, context ...redis.Cmdable) (int64, error) {
	if redisClient == nil {
		return 0, ErrDBNotInit
	}

	v, err := getCmdable(context...).Del(ctx, key).Result()
	return v, err
}

// Exist checks if a key exists
func Exist(key string, context ...redis.Cmdable) (int64, error) {
	if redisClient == nil {
		return 0, ErrDBNotInit
	}

	return getCmdable(context...).Exists(ctx, serialKey(key)).Result()
}

// Increase increments the value of a key
func Increase(key string, context ...redis.Cmdable) (int64, error) {
	if redisClient == nil {
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

// Decrease decrements the value of a key
func Decrease(key string, context ...redis.Cmdable) (int64, error) {
	if redisClient == nil {
		return 0, ErrDBNotInit
	}

	return getCmdable(context...).Decr(ctx, serialKey(key)).Result()
}

// SetExpire sets the expiration time for a key
func SetExpire(key string, time time.Duration, context ...redis.Cmdable) error {
	if redisClient == nil {
		return ErrDBNotInit
	}

	return getCmdable(context...).Expire(ctx, serialKey(key), time).Err()
}

// SetMapField sets hash fields
func SetMapField(key string, v map[string]any, context ...redis.Cmdable) error {
	if redisClient == nil {
		return ErrDBNotInit
	}

	return getCmdable(context...).HMSet(ctx, serialKey(key), v).Err()
}

// SetMapOneField sets a single hash field
func SetMapOneField(key string, field string, value any, context ...redis.Cmdable) error {
	if redisClient == nil {
		return ErrDBNotInit
	}

	if _, ok := value.(string); !ok {
		value = parser.MarshalJson(value)
	}

	return getCmdable(context...).HSet(ctx, serialKey(key), field, value).Err()
}

// GetMapField retrieves a hash field
func GetMapField[T any](key string, field string, context ...redis.Cmdable) (*T, error) {
	if redisClient == nil {
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

// GetMapFieldString retrieves a hash field as string
func GetMapFieldString(key string, field string, context ...redis.Cmdable) (string, error) {
	if redisClient == nil {
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

// DelMapField deletes a hash field
func DelMapField(key string, field string, context ...redis.Cmdable) error {
	if redisClient == nil {
		return ErrDBNotInit
	}

	return getCmdable(context...).HDel(ctx, serialKey(key), field).Err()
}

// GetMap retrieves the entire hash
func GetMap[V any](key string, context ...redis.Cmdable) (map[string]V, error) {
	if redisClient == nil {
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

// ScanKeys scans for matching keys
func ScanKeys(match string, context ...redis.Cmdable) ([]string, error) {
	if redisClient == nil {
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

// ScanKeysAsync asynchronously scans for matching keys
func ScanKeysAsync(match string, fn func([]string) error, context ...redis.Cmdable) error {
	if redisClient == nil {
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

// ScanMap scans for matching hash fields
func ScanMap[V any](key string, match string, context ...redis.Cmdable) (map[string]V, error) {
	if redisClient == nil {
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

// ScanMapAsync asynchronously scans for matching hash fields
func ScanMapAsync[V any](key string, match string, fn func(map[string]V) error, context ...redis.Cmdable) error {
	if redisClient == nil {
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

// SetNX sets a key-value pair only if the key does not exist
func SetNX[T any](key string, value T, expire time.Duration, context ...redis.Cmdable) (bool, error) {
	if redisClient == nil {
		return false, ErrDBNotInit
	}

	// Serialize value
	bytes, err := parser.MarshalCBOR(value)
	if err != nil {
		return false, err
	}

	return getCmdable(context...).SetNX(ctx, serialKey(key), bytes, expire).Result()
}

var (
	ErrLockTimeout = errors.New("lock timeout")
)

// Lock acquires a lock
func Lock(key string, expire time.Duration, tryLockTimeout time.Duration, context ...redis.Cmdable) error {
	if redisClient == nil {
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

// Unlock releases a lock
func Unlock(key string, context ...redis.Cmdable) error {
	if redisClient == nil {
		return ErrDBNotInit
	}

	return getCmdable(context...).Del(ctx, serialKey(key)).Err()
}

// Expire 设置过期时间
func Expire(key string, time time.Duration, context ...redis.Cmdable) (bool, error) {
	if redisClient == nil {
		return false, ErrDBNotInit
	}

	return getCmdable(context...).Expire(ctx, serialKey(key), time).Result()
}

// Transaction executes a transaction
func Transaction(fn func(redis.Pipeliner) error) error {
	if redisClient == nil {
		return ErrDBNotInit
	}

	// Determine client type
	switch client := redisClient.(type) {
	case *redis.Client:
		// Standalone Redis supports atomic transactions
		return client.Watch(ctx, func(tx *redis.Tx) error {
			_, err := tx.TxPipelined(ctx, func(p redis.Pipeliner) error {
				return fn(p)
			})
			if err == redis.Nil {
				return nil
			}
			return err
		})
	case *redis.ClusterClient:
		// In cluster mode, use pipeline without atomicity guarantees
		pipe := client.Pipeline()
		err := fn(pipe)
		if err != nil {
			return err
		}

		_, err = pipe.Exec(ctx)
		if err == redis.Nil {
			return nil
		}
		return err
	default:
		return fmt.Errorf("unknown redis client type: %T", client)
	}
}

// Publish publishes a message to a channel
func Publish(channel string, message any, context ...redis.Cmdable) error {
	if redisClient == nil {
		return ErrDBNotInit
	}

	if _, ok := message.(string); !ok {
		message = parser.MarshalJson(message)
	}

	return getCmdable(context...).Publish(ctx, channel, message).Err()
}

// Subscribe subscribes to a channel
func Subscribe[T any](channel string) (<-chan T, func()) {
	ch := make(chan T)
	connectionEstablished := make(chan bool)

	if redisClient == nil {
		close(ch)
		close(connectionEstablished)
		return ch, func() {}
	}

	// Get PubSub based on different client types
	var pubsub *redis.PubSub
	switch client := redisClient.(type) {
	case *redis.Client:
		pubsub = client.Subscribe(ctx, channel)
	case *redis.ClusterClient:
		pubsub = client.Subscribe(ctx, channel)
	default:
		log.Error("unknown redis client type: %T", client)
		close(ch)
		close(connectionEstablished)
		return ch, func() {}
	}

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

	// Wait for connection to be established
	<-connectionEstablished

	return ch, func() {
		pubsub.Close()
	}
}
