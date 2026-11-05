package cache

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
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
	redisDiscoveryMaxAttempts   = 12
	redisDiscoveryRetryWait     = 500 * time.Millisecond
	// Grace period before closing a replaced client so in-flight commands can finish.
	redisClientRetireDelay = 30 * time.Second
)

var errNoWritableRedisMaster = errors.New("no writable redis master discovered via sentinel")

// discoverWritableMasterHook is set in tests to simulate promotion lag or inconsistent sentinels.
var discoverWritableMasterHook func([]string, sentinelDiscoveryOptions) (string, map[string]string, error)

// rediscoverSentinelClientHook is set in tests to stub runtime rediscovery.
var rediscoverSentinelClientHook func() error

var (
	nodeRoleFn               = redisRole
	nodeWriteFn              = probeWrite
	sentinelReportsMasterFn  = sentinelReportsMaster

	sentinelRuntimeMu      sync.Mutex
	sentinelRuntime      *sentinelRuntimeConfig
	sentinelRediscoverMu sync.Mutex
	// Non-zero while openSentinelFailoverClient runs pre-hook validation (write probe).
	sentinelClientOpening atomic.Int32
)

type sentinelRuntimeConfig struct {
	sentinels        []string
	masterName       string
	creds            RedisCredentials
	sentinelUsername string
	sentinelPassword string
	useSsl           bool
	db               int
	socketTimeout    float64
	tlsConf          *tls.Config
}

func (c *sentinelRuntimeConfig) discoveryOptions() sentinelDiscoveryOptions {
	return sentinelDiscoveryOptions{
		masterName:       c.masterName,
		creds:            c.creds,
		sentinelUsername: c.sentinelUsername,
		sentinelPassword: c.sentinelPassword,
		useSsl:           c.useSsl,
		db:               c.db,
		socketTimeout:    c.socketTimeout,
		tlsConf:          c.tlsConf,
	}
}

func (c *sentinelRuntimeConfig) failoverOptions(sentinelAddrs []string) *redis.FailoverOptions {
	opts := &redis.FailoverOptions{
		MasterName:                   c.masterName,
		SentinelAddrs:                sentinelAddrs,
		Username:                     c.creds.Username,
		Password:                     c.creds.Password,
		DB:                           c.db,
		SentinelUsername:             c.sentinelUsername,
		SentinelPassword:             c.sentinelPassword,
		StreamingCredentialsProvider: c.creds.CredentialProvider,
		MaxRetries:                   5,
		MinRetryBackoff:              200 * time.Millisecond,
	}
	if c.useSsl {
		if c.tlsConf != nil {
			opts.TLSConfig = c.tlsConf
		} else {
			opts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		}
	}
	if c.socketTimeout > 0 {
		opts.DialTimeout = time.Duration(c.socketTimeout * float64(time.Second))
	}
	return opts
}

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
	return getRedisOptions(addr, o.creds, o.useSsl, o.db, o.tlsConf)
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

func sentinelMasterVotes(sentinels []string, masterName string, o sentinelDiscoveryOptions) map[string]int {
	votes := map[string]int{}
	for _, sentinelAddr := range sentinels {
		sentinel := redis.NewSentinelClient(o.sentinelClientOptions(sentinelAddr))
		parts, err := sentinel.GetMasterAddrByName(ctx, masterName).Result()
		_ = sentinel.Close()
		if err != nil || len(parts) != 2 {
			continue
		}
		addr := joinHostPort(parts[0], parts[1])
		votes[addr]++
	}
	return votes
}

func sentinelReportsMaster(sentinelAddr, masterName, masterAddr string, o sentinelDiscoveryOptions) bool {
	sentinel := redis.NewSentinelClient(o.sentinelClientOptions(sentinelAddr))
	parts, err := sentinel.GetMasterAddrByName(ctx, masterName).Result()
	_ = sentinel.Close()
	if err != nil || len(parts) != 2 {
		return false
	}
	return joinHostPort(parts[0], parts[1]) == masterAddr
}

// Reorders sentinels so those reporting writableMaster are tried first by FailoverClient.
// The full sentinel list is preserved (no narrowing).
func sortSentinelsPreferringWritableMaster(
	sentinels []string,
	masterName string,
	writableMaster string,
	o sentinelDiscoveryOptions,
) []string {
	out := append([]string(nil), sentinels...)
	sort.SliceStable(out, func(i, j int) bool {
		ri := sentinelReportsMasterFn(out[i], masterName, writableMaster, o)
		rj := sentinelReportsMasterFn(out[j], masterName, writableMaster, o)
		if ri != rj {
			return ri
		}
		return out[i] < out[j]
	})
	return out
}

func sortCandidatesBySentinelVotes(candidates map[string]struct{}, votes map[string]int) []string {
	addrs := make([]string, 0, len(candidates))
	for addr := range candidates {
		addrs = append(addrs, addr)
	}
	sort.Slice(addrs, func(i, j int) bool {
		vi, vj := votes[addrs[i]], votes[addrs[j]]
		if vi != vj {
			return vi > vj
		}
		return addrs[i] < addrs[j]
	})
	return addrs
}

func tryWritableMasterFromCandidates(
	candidateAddrs []string,
	o sentinelDiscoveryOptions,
) (writableMaster string, roles map[string]string, err error) {
	roles = map[string]string{}
	for _, addr := range candidateAddrs {
		role, roleErr := nodeRoleFn(addr, o.redisClientOptions(addr))
		if roleErr != nil {
			roles[addr] = fmt.Sprintf("unreachable (%v)", roleErr)
			continue
		}
		roles[addr] = role
		if role != "master" {
			continue
		}
		if writeErr := nodeWriteFn(addr, o.redisClientOptions(addr)); writeErr != nil {
			roles[addr] = fmt.Sprintf("master (write failed: %v)", writeErr)
			continue
		}
		return addr, roles, nil
	}
	return "", roles, errNoWritableRedisMaster
}

func discoverWritableMaster(
	sentinels []string,
	o sentinelDiscoveryOptions,
) (writableMaster string, roles map[string]string, err error) {
	if discoverWritableMasterHook != nil {
		return discoverWritableMasterHook(sentinels, o)
	}
	candidates := map[string]struct{}{}
	for _, sentinelAddr := range sentinels {
		for _, addr := range collectSentinelCandidateAddrs(sentinelAddr, o) {
			candidates[addr] = struct{}{}
		}
	}
	votes := sentinelMasterVotes(sentinels, o.masterName, o)
	sorted := sortCandidatesBySentinelVotes(candidates, votes)
	return tryWritableMasterFromCandidates(sorted, o)
}

func discoverWritableMasterWithRetry(
	sentinels []string,
	o sentinelDiscoveryOptions,
) (writableMaster string, roles map[string]string, err error) {
	var lastRoles map[string]string
	for attempt := 0; attempt < redisDiscoveryMaxAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(redisDiscoveryRetryWait)
			log.Warn(
				"redis sentinel: retrying writable master discovery",
				"attempt", attempt+1,
				"max_attempts", redisDiscoveryMaxAttempts,
			)
		}
		writableMaster, roles, err = discoverWritableMaster(sentinels, o)
		lastRoles = roles
		if err == nil {
			return writableMaster, roles, nil
		}
	}
	return "", lastRoles, fmt.Errorf("%w; candidate roles: %v", errNoWritableRedisMaster, lastRoles)
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

type sentinelReadonlyHook struct{}

func (sentinelReadonlyHook) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

func (h sentinelReadonlyHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(cmdCtx context.Context, cmd redis.Cmder) error {
		err := next(cmdCtx, cmd)
		if err == nil || !redis.IsReadOnlyError(err) {
			return err
		}
		if sentinelClientOpening.Load() > 0 {
			return err
		}
		log.Warn("redis READONLY during command; rediscovering sentinel master", "cmd", cmd.Name())
		if rediscoverErr := rediscoverSentinelClient(); rediscoverErr != nil {
			return err
		}
		return retryRedisCommandAfterReadonly(cmdCtx, cmd, err)
	}
}

func (h sentinelReadonlyHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(cmdCtx context.Context, cmds []redis.Cmder) error {
		err := next(cmdCtx, cmds)
		if err == nil || !redis.IsReadOnlyError(err) {
			return err
		}
		if sentinelClientOpening.Load() > 0 {
			return err
		}
		log.Warn("redis READONLY during pipeline; rediscovering sentinel master")
		if rediscoverErr := rediscoverSentinelClient(); rediscoverErr != nil {
			return err
		}
		return retryRedisPipelineAfterReadonly(cmdCtx, cmds, err)
	}
}

func activeSentinelRedisClient() *redis.Client {
	c, _ := loadRedisClient().(*redis.Client)
	return c
}

func retryRedisCommandAfterReadonly(cmdCtx context.Context, cmd redis.Cmder, readonlyErr error) error {
	active := activeSentinelRedisClient()
	if active == nil {
		return readonlyErr
	}
	if err := active.Process(cmdCtx, cmd); err != nil {
		if redis.IsReadOnlyError(err) {
			return readonlyErr
		}
		return err
	}
	return nil
}

func retryRedisPipelineAfterReadonly(cmdCtx context.Context, cmds []redis.Cmder, readonlyErr error) error {
	for _, cmd := range cmds {
		if err := retryRedisCommandAfterReadonly(cmdCtx, cmd, readonlyErr); err != nil {
			return err
		}
	}
	return nil
}

func retireRedisClient(c redis.UniversalClient) {
	if c == nil {
		return
	}
	go func() {
		time.Sleep(redisClientRetireDelay)
		_ = c.Close()
	}()
}

func rediscoverSentinelClient() error {
	if rediscoverSentinelClientHook != nil {
		return rediscoverSentinelClientHook()
	}

	if !sentinelRediscoverMu.TryLock() {
		sentinelRediscoverMu.Lock()
		sentinelRediscoverMu.Unlock()
		return nil
	}
	defer sentinelRediscoverMu.Unlock()

	sentinelRuntimeMu.Lock()
	cfg := sentinelRuntime
	sentinelRuntimeMu.Unlock()
	if cfg == nil {
		return errors.New("sentinel runtime not configured")
	}

	newClient, err := openSentinelFailoverClient(cfg)
	if err != nil {
		return err
	}

	old := swapRedisClient(newClient)
	retireRedisClient(old)
	return nil
}

func openSentinelFailoverClient(cfg *sentinelRuntimeConfig) (*redis.Client, error) {
	o := cfg.discoveryOptions()
	writableMaster, roles, err := discoverWritableMasterWithRetry(cfg.sentinels, o)
	if err != nil {
		return nil, err
	}
	log.Info(
		"redis sentinel: using writable master reported by sentinel quorum",
		"master", writableMaster,
		"roles", roles,
	)

	sentinelAddrs := sortSentinelsPreferringWritableMaster(cfg.sentinels, cfg.masterName, writableMaster, o)
	c := redis.NewFailoverClient(cfg.failoverOptions(sentinelAddrs))
	_ = redisotel.InstrumentTracing(c, redisotel.WithTracerProvider(gootel.GetTracerProvider()))

	sentinelClientOpening.Add(1)
	defer sentinelClientOpening.Add(-1)

	if _, err := c.Ping(ctx).Result(); err != nil {
		_ = c.Close()
		return nil, err
	}
	if err := ensureRedisWritable(c); err != nil {
		_ = c.Close()
		return nil, err
	}
	c.AddHook(sentinelReadonlyHook{})
	return c, nil
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

	cfg := &sentinelRuntimeConfig{
		sentinels:        append([]string(nil), sentinels...),
		masterName:       masterName,
		creds:            creds,
		sentinelUsername: sentinelUsername,
		sentinelPassword: sentinelPassword,
		useSsl:           useSsl,
		db:               db,
		socketTimeout:    socketTimeout,
		tlsConf:          tlsConf,
	}

	sentinelRuntimeMu.Lock()
	sentinelRuntime = cfg
	sentinelRuntimeMu.Unlock()

	c, err := openSentinelFailoverClient(cfg)
	if err != nil {
		return err
	}
	old := swapRedisClient(c)
	retireRedisClient(old)
	return nil
}
