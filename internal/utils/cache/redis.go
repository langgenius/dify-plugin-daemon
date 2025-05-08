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

// RedisInterface 是 Redis 客户端的抽象接口
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

// StandaloneRedis 包装单实例 Redis 客户端
type StandaloneRedis struct {
	*redis.Client
}

// ClusterRedis 包装集群 Redis 客户端
type ClusterRedis struct {
	*redis.ClusterClient
}

// Pipeline 包装 Pipeline 方法，确保 ClusterRedis 实现接口
func (c *ClusterRedis) Pipeline() redis.Pipeliner {
	return c.ClusterClient.Pipeline()
}

// 创建 Redis 单实例配置选项
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

// InitStandaloneClient 初始化单实例 Redis 客户端（内部使用）
func InitStandaloneClient(addr, username, password string, useSsl bool, db int) (redis.Cmdable, error) {
	opts := getRedisOptions(addr, username, password, useSsl, db)
	client := redis.NewClient(opts)

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	return client, nil
}

// InitClusterClient 初始化集群 Redis 客户端（内部使用）
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

// InitSentinelClient 初始化哨兵模式 Redis 客户端（内部使用）
func InitSentinelClient(
	masterName string,
	sentinelAddrs []string,
	username string, // Redis 主从服务器的用户名
	password string, // Redis 主从服务器的密码
	sentinelUsername string, // 连接哨兵服务器的用户名
	sentinelPassword string, // 连接哨兵服务器的密码
	useSsl bool,
	db int,
	socketTimeout float64,
) (redis.Cmdable, error) {
	opts := &redis.FailoverOptions{
		MasterName:       masterName,
		SentinelAddrs:    sentinelAddrs,
		Username:         username,         // Redis 主从服务器的用户名
		Password:         password,         // Redis 主从服务器的密码
		SentinelUsername: sentinelUsername, // 连接哨兵服务器的用户名
		SentinelPassword: sentinelPassword, // 连接哨兵服务器的密码
		DB:               db,
	}

	// 设置套接字超时
	if socketTimeout > 0 {
		opts.DialTimeout = time.Duration(socketTimeout * float64(time.Second))
	}

	if useSsl {
		opts.TLSConfig = &tls.Config{}
	}

	client := redis.NewFailoverClient(opts)

	// 打印客户端类型，用于调试
	log.Info("Redis failover client type: %T", client)

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	return client, nil
}

// InitRedisClient 初始化单实例 Redis 客户端
func InitRedisClient(addr, username, password string, useSsl bool, db int) error {
	client, err := InitStandaloneClient(addr, username, password, useSsl, db)
	if err != nil {
		return err
	}

	redisClient = client
	return nil
}

// InitRedisClusterClient 初始化集群 Redis 客户端
func InitRedisClusterClient(addrs []string, password string, useSsl bool) error {
	client, err := InitClusterClient(addrs, password, useSsl)
	if err != nil {
		return err
	}

	redisClient = client
	return nil
}

// InitRedisSentinelClient 初始化哨兵模式 Redis 客户端
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

// Close 关闭 Redis 客户端
func Close() error {
	if redisClient == nil {
		return ErrDBNotInit
	}

	// 根据不同的客户端类型进行关闭
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

// Store 存储键值对
func Store(key string, value any, time time.Duration, context ...redis.Cmdable) error {
	return store(serialKey(key), value, time, context...)
}

// store 存储键值对，不使用序列化 key
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

// Get 获取键值
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

// GetString 获取字符串类型的键值
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

// Del 删除键
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

// Exist 检查键是否存在
func Exist(key string, context ...redis.Cmdable) (int64, error) {
	if redisClient == nil {
		return 0, ErrDBNotInit
	}

	return getCmdable(context...).Exists(ctx, serialKey(key)).Result()
}

// Increase 增加键的值
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

// Decrease 减少键的值
func Decrease(key string, context ...redis.Cmdable) (int64, error) {
	if redisClient == nil {
		return 0, ErrDBNotInit
	}

	return getCmdable(context...).Decr(ctx, serialKey(key)).Result()
}

// SetExpire 设置键的过期时间
func SetExpire(key string, time time.Duration, context ...redis.Cmdable) error {
	if redisClient == nil {
		return ErrDBNotInit
	}

	return getCmdable(context...).Expire(ctx, serialKey(key), time).Err()
}

// SetMapField 设置哈希表字段
func SetMapField(key string, v map[string]any, context ...redis.Cmdable) error {
	if redisClient == nil {
		return ErrDBNotInit
	}

	return getCmdable(context...).HMSet(ctx, serialKey(key), v).Err()
}

// SetMapOneField 设置哈希表单个字段
func SetMapOneField(key string, field string, value any, context ...redis.Cmdable) error {
	if redisClient == nil {
		return ErrDBNotInit
	}

	if _, ok := value.(string); !ok {
		value = parser.MarshalJson(value)
	}

	return getCmdable(context...).HSet(ctx, serialKey(key), field, value).Err()
}

// GetMapField 获取哈希表字段
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

// GetMapFieldString 获取哈希表字段字符串
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

// DelMapField 删除哈希表字段
func DelMapField(key string, field string, context ...redis.Cmdable) error {
	if redisClient == nil {
		return ErrDBNotInit
	}

	return getCmdable(context...).HDel(ctx, serialKey(key), field).Err()
}

// GetMap 获取整个哈希表
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

// ScanKeys 扫描匹配的键
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

// ScanKeysAsync 异步扫描匹配的键
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

// ScanMap 扫描匹配的哈希表字段
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

// ScanMapAsync 异步扫描匹配的哈希表字段
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

// SetNX 设置键值对，仅当键不存在时
func SetNX[T any](key string, value T, expire time.Duration, context ...redis.Cmdable) (bool, error) {
	if redisClient == nil {
		return false, ErrDBNotInit
	}

	// 序列化值
	bytes, err := parser.MarshalCBOR(value)
	if err != nil {
		return false, err
	}

	return getCmdable(context...).SetNX(ctx, serialKey(key), bytes, expire).Result()
}

var (
	ErrLockTimeout = errors.New("lock timeout")
)

// Lock 加锁
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

// Unlock 解锁
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

// Transaction 执行事务
func Transaction(fn func(redis.Pipeliner) error) error {
	if redisClient == nil {
		return ErrDBNotInit
	}

	// 判断客户端类型
	switch client := redisClient.(type) {
	case *redis.Client:
		// 单实例 Redis 支持原子性事务
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
		// 集群模式下使用管道，不保证原子性
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

// Publish 发布消息到频道
func Publish(channel string, message any, context ...redis.Cmdable) error {
	if redisClient == nil {
		return ErrDBNotInit
	}

	if _, ok := message.(string); !ok {
		message = parser.MarshalJson(message)
	}

	return getCmdable(context...).Publish(ctx, channel, message).Err()
}

// Subscribe 订阅频道
func Subscribe[T any](channel string) (<-chan T, func()) {
	ch := make(chan T)
	connectionEstablished := make(chan bool)

	if redisClient == nil {
		close(ch)
		close(connectionEstablished)
		return ch, func() {}
	}

	// 根据不同的客户端类型获取 PubSub
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

	// 等待连接建立
	<-connectionEstablished

	return ch, func() {
		pubsub.Close()
	}
}
