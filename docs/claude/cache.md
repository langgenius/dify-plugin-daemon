# Cache Operations

Redis-based caching system (`internal/utils/cache/`).

**Related Documentation:**
- [Stream Operations](stream.md) - Pub/sub streaming patterns
- [Database Operations](database.md) - Cache-aside pattern for DB results
- [Generic Types](generics.md) - Type-safe cache operations

## Basic Operations

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
    addr,      // "localhost:6379"
    username,  // optional
    password,  
    useSsl,    // bool
    db,        // database number
)
```