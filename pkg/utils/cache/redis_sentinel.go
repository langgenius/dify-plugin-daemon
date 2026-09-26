package cache

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	gootel "go.opentelemetry.io/otel"
)

const (
	redisMasterWriteProbeKey    = "__plugin_daemon_master_write_probe__"
	redisMasterWriteProbeTTL    = 5 * time.Second
	redisMasterWriteMaxAttempts = 12
	redisMasterWriteRetryWait   = 500 * time.Millisecond
)

var errNoWritableRedisMaster = errors.New("no writable redis master discovered via sentinel")

type sentinelDiscoveryOptions struct {
	masterName       string
	creds            RedisCredentials
	sentinelUsername string
	sentinelPassword string
	useSsl           bool
	db               int
	socketTimeout    float64
	tlsConf          *tls.Config
}

func (o sentinelDiscoveryOptions) sentinelClientOptions(sentinelAddr string) *redis.Options {
	opts := &redis.Options{
		Addr:     sentinelAddr,
		Username: o.sentinelUsername,
		Password: o.sentinelPassword,
	}
	if o.socketTimeout > 0 {
		opts.DialTimeout = time.Duration(o.socketTimeout * float64(time.Second))
	}
	if o.useSsl {
		if o.tlsConf != nil {
			opts.TLSConfig = o.tlsConf
		} else {
			opts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		}
	}
	return opts
}

func (o sentinelDiscoveryOptions) redisClientOptions(addr string) *redis.Options {
	opts := getRedisOptions(addr, o.creds, o.useSsl, o.db, o.tlsConf)
	return opts
}

func joinHostPort(host, port string) string {
	if strings.Contains(host, ":") && !strings.HasPrefix(host, "[") {
		return net.JoinHostPort(host, port)
	}
	return net.JoinHostPort(host, port)
}

func redisRole(addr string, opts *redis.Options) (string, error) {
	cli := redis.NewClient(opts)
	defer cli.Close()

	val, err := cli.Do(ctx, "ROLE").Result()
	if err != nil {
		return "", err
	}
	parts, ok := val.([]any)
	if !ok || len(parts) == 0 {
		return "", fmt.Errorf("unexpected ROLE response from %s", addr)
	}
	role, ok := parts[0].(string)
	if !ok {
		return "", fmt.Errorf("unexpected ROLE role from %s", addr)
	}
	return role, nil
}

func probeWrite(addr string, opts *redis.Options) error {
	cli := redis.NewClient(opts)
	defer cli.Close()

	probeKey := serialKey(redisMasterWriteProbeKey)
	if err := cli.Set(ctx, probeKey, "1", redisMasterWriteProbeTTL).Err(); err != nil {
		return err
	}
	_ = cli.Del(ctx, probeKey).Err()
	return nil
}

func collectSentinelCandidateAddrs(sentinelAddr string, o sentinelDiscoveryOptions) []string {
	sentinel := redis.NewSentinelClient(o.sentinelClientOptions(sentinelAddr))
	defer sentinel.Close()

	seen := map[string]struct{}{}
	add := func(host, port string) {
		if host == "" || port == "" {
			return
		}
		addr := joinHostPort(host, port)
		seen[addr] = struct{}{}
	}

	if parts, err := sentinel.GetMasterAddrByName(ctx, o.masterName).Result(); err == nil && len(parts) == 2 {
		add(parts[0], parts[1])
	}
	if replicas, err := sentinel.Replicas(ctx, o.masterName).Result(); err == nil {
		for _, replica := range replicas {
			add(replica["ip"], replica["port"])
		}
	}

	out := make([]string, 0, len(seen))
	for addr := range seen {
		out = append(out, addr)
	}
	return out
}

func discoverWritableMaster(
	sentinels []string,
	o sentinelDiscoveryOptions,
) (writableMaster string, roles map[string]string, err error) {
	roles = map[string]string{}
	candidates := map[string]struct{}{}

	for _, sentinelAddr := range sentinels {
		for _, addr := range collectSentinelCandidateAddrs(sentinelAddr, o) {
			candidates[addr] = struct{}{}
		}
	}

	for addr := range candidates {
		role, roleErr := redisRole(addr, o.redisClientOptions(addr))
		if roleErr != nil {
			roles[addr] = fmt.Sprintf("unreachable (%v)", roleErr)
			continue
		}
		roles[addr] = role
		if role != "master" {
			continue
		}
		if writeErr := probeWrite(addr, o.redisClientOptions(addr)); writeErr != nil {
			roles[addr] = fmt.Sprintf("master (write failed: %v)", writeErr)
			continue
		}
		return addr, roles, nil
	}

	return "", roles, errNoWritableRedisMaster
}

func sentinelsAgreeOnMaster(sentinels []string, masterName, masterAddr string, o sentinelDiscoveryOptions) []string {
	matched := make([]string, 0, len(sentinels))
	for _, sentinelAddr := range sentinels {
		sentinel := redis.NewSentinelClient(o.sentinelClientOptions(sentinelAddr))
		parts, err := sentinel.GetMasterAddrByName(ctx, masterName).Result()
		_ = sentinel.Close()
		if err != nil || len(parts) != 2 {
			continue
		}
		if joinHostPort(parts[0], parts[1]) == masterAddr {
			matched = append(matched, sentinelAddr)
		}
	}
	return matched
}

func ensureRedisWritable(c redis.Cmdable) error {
	probeKey := serialKey(redisMasterWriteProbeKey)
	var lastErr error

	for attempt := 0; attempt < redisMasterWriteMaxAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(redisMasterWriteRetryWait)
		}
		err := c.Set(ctx, probeKey, "1", redisMasterWriteProbeTTL).Err()
		if err == nil {
			_ = c.Del(ctx, probeKey).Err()
			return nil
		}
		lastErr = err
		if !redis.IsReadOnlyError(err) {
			return fmt.Errorf("redis write probe failed: %w", err)
		}
		log.Warn(
			"redis node is read-only; waiting for sentinel to promote the master",
			"attempt", attempt+1,
			"max_attempts", redisMasterWriteMaxAttempts,
		)
	}

	return fmt.Errorf(
		"redis master is not writable after %d attempts (check sentinel master address with SENTINEL get-master-addr-by-name and INFO replication on that host): %w",
		redisMasterWriteMaxAttempts,
		lastErr,
	)
}

func initRedisSentinelClientWithDiscovery(
	sentinels []string,
	masterName string,
	creds RedisCredentials,
	sentinelUsername, sentinelPassword string,
	useSsl bool,
	db int,
	socketTimeout float64,
	tlsConf *tls.Config,
) error {
	if len(sentinels) == 0 {
		return errors.New("redis sentinel addresses are required")
	}
	if strings.TrimSpace(masterName) == "" {
		return errors.New("redis sentinel service name is required")
	}

	discovery := sentinelDiscoveryOptions{
		masterName:       masterName,
		creds:            creds,
		sentinelUsername: sentinelUsername,
		sentinelPassword: sentinelPassword,
		useSsl:           useSsl,
		db:               db,
		socketTimeout:    socketTimeout,
		tlsConf:          tlsConf,
	}

	writableMaster, roles, discoverErr := discoverWritableMaster(sentinels, discovery)
	if discoverErr != nil {
		return fmt.Errorf("%w; candidate roles: %v", discoverErr, roles)
	}

	filteredSentinels := sentinelsAgreeOnMaster(sentinels, masterName, writableMaster, discovery)
	if len(filteredSentinels) > 0 {
		log.Info(
			"redis sentinel: using sentinels that report the writable master",
			"master", writableMaster,
			"sentinels", len(filteredSentinels),
		)
		sentinels = filteredSentinels
	} else {
		log.Warn(
			"redis sentinel: no sentinel reports the writable master; using all configured sentinels",
			"master", writableMaster,
			"roles", roles,
		)
	}

	opts := &redis.FailoverOptions{
		MasterName:                   masterName,
		SentinelAddrs:                sentinels,
		Username:                     creds.Username,
		Password:                     creds.Password,
		DB:                           db,
		SentinelUsername:             sentinelUsername,
		SentinelPassword:             sentinelPassword,
		StreamingCredentialsProvider: creds.CredentialProvider,
		MaxRetries:                   5,
		MinRetryBackoff:              200 * time.Millisecond,
	}

	if useSsl {
		if tlsConf != nil {
			opts.TLSConfig = tlsConf
		} else {
			opts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		}
	}
	if socketTimeout > 0 {
		opts.DialTimeout = time.Duration(socketTimeout * float64(time.Second))
	}

	client = redis.NewFailoverClient(opts)
	_ = redisotel.InstrumentTracing(client, redisotel.WithTracerProvider(gootel.GetTracerProvider()))

	if _, err := client.Ping(ctx).Result(); err != nil {
		return err
	}
	return ensureRedisWritable(client)
}
