# Cache Operations

Redis-based caching system (`internal/utils/cache/`).

**Related Documentation:**
- [Stream Operations](stream.md) - Pub/sub streaming patterns
- [Database Operations](database.md) - Cache-aside pattern for DB results
- [Generic Types](generics.md) - Type-safe cache operations

## Basic Operations

All cache helpers accept logical names and apply the configured Redis prefix internally.
By default the daemon uses `plugin_daemon`, and operators can override it with
`REDIS_KEY_PREFIX`.

```go
// Store and retrieve
cache.Store("key", value, time.Minute*30)
val, err := cache.Get[Type]("key")
cache.Del("key")

// Check existence
exists, _ := cache.Exist("key")
```

## Map Operations

For cluster state management:

```go
// Set map field
cache.SetMapOneField(CLUSTER_STATUS_KEY, nodeId, nodeStatus)

// Get single field
node, err := cache.GetMapField[node](CLUSTER_STATUS_KEY, nodeId)

// Get entire map
nodes, err := cache.GetMap[node](CLUSTER_STATUS_KEY)

// Delete field
cache.DelMapField(CLUSTER_STATUS_KEY, nodeId)
```

## Pub/Sub Pattern

Pub/sub channels are prefixed with the same `REDIS_KEY_PREFIX` value as keys, so
callers should always pass logical channel names.

```go
// Publish event
cache.Publish(CHANNEL, event)

// Subscribe to channel
eventChan, cancel := cache.Subscribe[EventType](CHANNEL)
defer cancel()

for event := range eventChan {
    // Process event
}
```

## Distributed Locks

```go
// Acquire lock with timeout
acquired := cache.Lock(key, duration, timeout)
if acquired {
    defer cache.Unlock(key)
    // Critical section
}
```

## Auto-type Operations

Helper functions in `redis_auto_type.go`:

```go
// Get with automatic getter fallback
value := cache.AutoGetWithGetter(key, func() (*Type, error) {
    return fetchFromDB()
}, ttl)

// Auto delete
cache.AutoDelete[Type](key)
```

## Session Management Example

```go
// Store session
sessionKey := fmt.Sprintf("session_info:%s", id)
cache.Store(sessionKey, session, time.Minute*30)

// Retrieve session
session, err := cache.Get[Session](sessionKey)
if err == cache.ErrNotFound {
    // Session expired or not found
}
```

## Configuration

Initialize Redis client in main:

```go
cache.InitRedisClient(
    addr,  // "localhost:6379"
    cache.RedisCredentials{
        Username: username,  // optional
        Password: password,
        CredentialProvider: nil,  // optional StreamingCredentialsProvider
    },
    useSsl,  // bool
    db,      // database number
    nil,     // *tls.Config (optional)
)
```

### Azure Managed Identity Authentication

Azure Entra ID authentication is supported via the `REDIS_USE_AZURE_MANAGED_IDENTITY` environment variable. When enabled, it uses `DefaultAzureCredential` for token-based authentication instead of static username/password credentials.

**Important:** Azure Managed Redis only supports database 0, so `REDIS_DB` must be set to 0 when using this feature.

```go
// Create Azure credentials provider
provider, err := cache.NewAzureEntraIDCredentialsProvider()
if err != nil {
    log.Fatal("failed to create Azure credentials provider:", err)
}

cache.InitRedisClient(
    addr,
    cache.RedisCredentials{
        CredentialProvider: provider,  // takes precedence over Username/Password
    },
    useSsl,
    db,
    nil,
)
```

When `CredentialProvider` is set, it takes precedence over static `Username` and `Password` fields.

### Redis Naming Behavior

- `REDIS_KEY_PREFIX` defaults to `plugin_daemon`
- applies to keys and pub/sub channels managed by this package
- changing the prefix switches the active Redis namespace and does not migrate old data
- logical keys that contain Redis Cluster hash tags such as `{remote:key:manager}` keep
  those tags intact because the prefix is prepended to the full logical key
